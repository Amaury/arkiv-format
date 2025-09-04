package arkivformat

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ensureParents creates intermediate directories for a path (mkdir -p).
func ensureParents(path string) error {
	return os.MkdirAll(filepath.Dir(path), 0o755)
}

// Extract restores files under dest for entries matching optional prefixes.
// It loads prefix+index lazily, then performs a second pass over the tar
// to create objects and write data. File metadata is applied AFTER writing
// to ensure modes take effect even with restrictive umask.
func (a *ArchiveReader) Extract(dest string, prefixes []string) error {
	// Ensure prefix and index are ready.
	if err := a.ensureLoaded(); err != nil {
		return err
	}

	// Build the subset of entries to extract.
	wanted := make([]IndexEntry, 0, len(a.index.Entries))
	for _, e := range a.index.Entries {
		if matchesPrefix(e.PathRaw, prefixes) {
			wanted = append(wanted, e)
		}
	}
	if len(wanted) == 0 {
		return nil
	}

	// Mapping helpers for meta and data names.
	targetNameHashes := make(map[string]IndexEntry, len(wanted))
	dataNeeds := make(map[string][]IndexEntry)
	regMetaByPath := make(map[string]*tar.Header)

	for _, e := range wanted {
		hName := computeNameHash(a.prefixB64, e.PathRaw)
		metaName := filepath.ToSlash(filepath.Join("meta", hName+".tar.zst.aes"))
		targetNameHashes[metaName] = e
		if e.HashData != "" {
			dataName := filepath.ToSlash(filepath.Join("data", e.HashData+".zst.aes"))
			dataNeeds[dataName] = append(dataNeeds[dataName], e)
		}
	}

	// Helper to convert raw stored path to output filesystem path.
	toOutPath := func(raw string) string {
		p := strings.ReplaceAll(raw, "\\", "\\\\")
		p = strings.ReplaceAll(p, "\"", "\\\"")
		return filepath.Join(dest, p)
	}

	// Ensure destination exists.
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}

	// Second pass: iterate members and act on meta/data.
	f, err := os.Open(a.path)
	if err != nil {
		return err
	}
	defer f.Close()

	tr := tar.NewReader(f)

	// Skip magic.zst and prefix.zst.aes.
	if _, err := tr.Next(); err != nil {
		return err
	}
	if _, err := tr.Next(); err != nil {
		return err
	}

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Process meta entries for wanted paths.
		if e, ok := targetNameHashes[hdr.Name]; ok {
			dr, err := OpenSSLDecryptReader(tr, a.password)
			if err != nil {
				return err
			}
			zdec, err := NewZstdDecoder(dr)
			if err != nil {
				return err
			}
			mtr := tar.NewReader(zdec)
			mh, err := mtr.Next()
			zdec.Close()
			if err != nil {
				return err
			}

			outPath := toOutPath(e.PathRaw)
			switch mh.Typeflag {
			case tar.TypeDir:
				if err := os.MkdirAll(outPath, os.FileMode(mh.Mode)); err != nil {
					return err
				}
				_ = chownBestEffort(outPath, mh.Uid, mh.Gid)
				_ = os.Chtimes(outPath, time.Now(), mh.ModTime)

			case tar.TypeSymlink:
				if err := ensureParents(outPath); err != nil {
					return err
				}
				if err := os.Symlink(mh.Linkname, outPath); err != nil {
					return err
				}
				_ = chownBestEffort(outPath, mh.Uid, mh.Gid)

			case tar.TypeFifo:
				if err := ensureParents(outPath); err != nil {
					return err
				}
				if err := mkfifo(outPath, uint32(mh.Mode)); err != nil {
					return err
				}
				_ = chownBestEffort(outPath, mh.Uid, mh.Gid)
				_ = os.Chtimes(outPath, time.Now(), mh.ModTime)

			case tar.TypeReg:
				// Regular files: defer metadata application after data write.
				regMetaByPath[e.PathRaw] = mh
			}
			continue
		}

		// Process data chunks for wanted regular files.
		if entries, ok := dataNeeds[hdr.Name]; ok {
			dr, err := OpenSSLDecryptReader(tr, a.password)
			if err != nil {
				return err
			}
			zdec, err := NewZstdDecoder(dr)
			if err != nil {
				return err
			}
			for _, e := range entries {
				mh := regMetaByPath[e.PathRaw]
				if mh == nil {
					zdec.Close()
					return fmt.Errorf("missing meta for regular file %s", e.PathRaw)
				}
				outPath := toOutPath(e.PathRaw)
				if err := ensureParents(outPath); err != nil {
					zdec.Close()
					return err
				}
				out, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(mh.Mode))
				if err != nil {
					zdec.Close()
					return err
				}
				if _, err := io.Copy(out, zdec); err != nil {
					out.Close()
					zdec.Close()
					return err
				}
				out.Close()
				_ = os.Chmod(outPath, os.FileMode(mh.Mode))
				_ = chownBestEffort(outPath, mh.Uid, mh.Gid)
				_ = os.Chtimes(outPath, time.Now(), mh.ModTime)
			}
			zdec.Close()
			continue
		}
	}
	return nil
}

