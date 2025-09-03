package arkivformat

import (
	"archive/tar"
	"bytes"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Create writes a new Arkiv archive at writer.path using the provided
// input file system paths. It writes members in this order:
//   magic.zst → prefix.zst.aes → meta/* and data/* → index.zst.aes (last)
// It strictly adheres to the Arkiv format for full compatibility.
func (w *ArchiveWriter) Create(inputs []string) error {
	// Create (or truncate) the destination archive file.
	f, err := os.Create(w.path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Prepare tar writer for the outer container.
	tw := tar.NewWriter(f)
	defer tw.Close()

	// --- Write magic.zst (zstd of "arkiv001", unencrypted) ---
	var magicBuf bytes.Buffer
	zwMagic, err := NewZstdEncoder(&magicBuf)
	if err != nil {
		return err
	}
	if _, err := zwMagic.Write([]byte(MagicString)); err != nil {
		zwMagic.Close()
		return err
	}
	if err := zwMagic.Close(); err != nil {
		return err
	}
	if err := tw.WriteHeader(&tar.Header{ Name: "magic.zst", Mode: 0644, Size: int64(magicBuf.Len()) }); err != nil {
		return err
	}
	if _, err := tw.Write(magicBuf.Bytes()); err != nil {
		return err
	}

	// --- Write prefix.zst.aes: 8 random bytes → zstd → OpenSSL enc ---
	prefixRaw := make([]byte, 8)
	if _, err := io.ReadFull(rand.Reader, prefixRaw); err != nil {
		return err
	}
	prefixB64 := base64.StdEncoding.EncodeToString(prefixRaw)

	var prefixEnc bytes.Buffer
	encW, err := OpenSSLEncryptWriter(&prefixEnc, w.password)
	if err != nil {
		return err
	}
	zwPrefix, err := NewZstdEncoder(encW)
	if err != nil {
		return err
	}
	if _, err := zwPrefix.Write(prefixRaw); err != nil {
		zwPrefix.Close()
		encW.Close()
		return err
	}
	if err := zwPrefix.Close(); err != nil {
		encW.Close()
		return err
	}
	if err := encW.Close(); err != nil {
		return err
	}
	if err := tw.WriteHeader(&tar.Header{ Name: "prefix.zst.aes", Mode: 0600, Size: int64(prefixEnc.Len()) }); err != nil {
		return err
	}
	if _, err := tw.Write(prefixEnc.Bytes()); err != nil {
		return err
	}

	// --- Walk inputs, collect paths (include directory itself, no symlink following) ---
	paths := make([]string, 0)
	for _, in := range inputs {
		in = filepath.Clean(in)
		err := filepath.WalkDir(in, func(p string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			// Always include the visited path.
			paths = append(paths, p)
			return nil
		})
		if err != nil {
			return err
		}
	}
	// Sort with C-locale byte ordering and de-duplicate exact paths.
	sort.Slice(paths, func(i, j int) bool {
		return strings.Compare(paths[i], paths[j]) < 0
	})
	uniq := make([]string, 0, len(paths))
	var last string
	for _, p := range paths {
		if p == last {
			continue
		}
		uniq = append(uniq, p)
		last = p
	}
	paths = uniq

	// Prepare the textual index and a set to avoid duplicate data writes.
	idx := Index{}
	dataWritten := make(map[string]bool)

	// --- Emit meta/* (and data/* for regular files) for each path ---
	for _, p := range paths {
		fi, err := os.Lstat(p)
		if err != nil {
			return err
		}

		// Detect file type and attributes.
		mode := fi.Mode()
		ft, linkname, err := classifyPath(p, fi)
		if err != nil {
			return err
		}

		// Build index entry (quoted path string and raw substring).
		quoted, raw := escapeForIndex(p)
		entry := IndexEntry{ PathRaw: raw, Quoted: quoted }

		// Compute HASH_NAME for the meta object.
		hName := computeNameHash(prefixB64, raw)
		metaName := filepath.ToSlash(filepath.Join("meta", hName+".tar.zst.aes"))

		// Create a one-entry tar carrying metadata only.
		var metaTar bytes.Buffer
		mtw := tar.NewWriter(&metaTar)
		hdr := &tar.Header{
			Name:    raw,                 // exact raw path between quotes
			Mode:    int64(fi.Mode().Perm()),
			Uid:     getUID(fi),
			Gid:     getGID(fi),
			ModTime: fi.ModTime().UTC(),  // store UTC
		}
		switch ft {
		case 'f':
			hdr.Typeflag = tar.TypeReg
			hdr.Size = 0 // metadata stub only
		case 'd':
			hdr.Typeflag = tar.TypeDir
		case 'l':
			hdr.Typeflag = tar.TypeSymlink
			hdr.Linkname = linkname
		case 'p':
			hdr.Typeflag = tar.TypeFifo
		default:
			return errors.New("unexpected file type")
		}
		if err := mtw.WriteHeader(hdr); err != nil {
			return err
		}
		if err := mtw.Close(); err != nil {
			return err
		}

		// Compress + encrypt the meta tar and write into the outer tar.
		var metaEnc bytes.Buffer
		encW, err := OpenSSLEncryptWriter(&metaEnc, w.password)
		if err != nil {
			return err
		}
		zwMeta, err := NewZstdEncoder(encW)
		if err != nil {
			encW.Close()
			return err
		}
		if _, err := zwMeta.Write(metaTar.Bytes()); err != nil {
			zwMeta.Close()
			encW.Close()
			return err
		}
		if err := zwMeta.Close(); err != nil {
			encW.Close()
			return err
		}
		if err := encW.Close(); err != nil {
			return err
		}
		if err := tw.WriteHeader(&tar.Header{ Name: metaName, Mode: 0600, Size: int64(metaEnc.Len()) }); err != nil {
			return err
		}
		if _, err := tw.Write(metaEnc.Bytes()); err != nil {
			return err
		}

		// For regular files, stream and write data/<HASH_DATA>.zst.aes once.
		if ft == 'f' {
			// Compute HASH_DATA while streaming raw file bytes through zstd+enc.
			h := sha512.New512_256()
			_, _ = h.Write([]byte(prefixB64))

			fData, err := os.Open(p)
			if err != nil {
				return err
			}

			var dataEnc bytes.Buffer
			encW, err := OpenSSLEncryptWriter(&dataEnc, w.password)
			if err != nil {
				fData.Close()
				return err
			}
			zwData, err := NewZstdEncoder(encW)
			if err != nil {
				encW.Close()
				fData.Close()
				return err
			}

			buf := make([]byte, 1<<20)
			for {
				n, er := fData.Read(buf)
				if n > 0 {
					_, _ = h.Write(buf[:n])
					if _, ew := zwData.Write(buf[:n]); ew != nil {
						zwData.Close()
						encW.Close()
						fData.Close()
						return ew
					}
				}
				if er == io.EOF {
					break
				}
				if er != nil {
					zwData.Close()
					encW.Close()
					fData.Close()
					return er
				}
			}
			fData.Close()
			if err := zwData.Close(); err != nil {
				encW.Close()
				return err
			}
			if err := encW.Close(); err != nil {
				return err
			}

			hData := hex.EncodeToString(h.Sum(nil))
			entry.HashData = hData

			if !dataWritten[hData] {
				dataWritten[hData] = true
				dataName := filepath.ToSlash(filepath.Join("data", hData+".zst.aes"))
				if err := tw.WriteHeader(&tar.Header{ Name: dataName, Mode: 0600, Size: int64(dataEnc.Len()) }); err != nil {
					return err
				}
				if _, err := tw.Write(dataEnc.Bytes()); err != nil {
					return err
				}
			}
		}

		// Add the entry to the textual index.
		idx.Entries = append(idx.Entries, entry)
	}

	// --- Finally, write index.zst.aes with sorted unique lines ---
	idxBytes := idx.Serialize()
	var idxEnc bytes.Buffer
	encW, err = OpenSSLEncryptWriter(&idxEnc, w.password)
	if err != nil {
		return err
	}
	zwIndex, err := NewZstdEncoder(encW)
	if err != nil {
		encW.Close()
		return err
	}
	if _, err := zwIndex.Write(idxBytes); err != nil {
		zwIndex.Close()
		encW.Close()
		return err
	}
	if err := zwIndex.Close(); err != nil {
		encW.Close()
		return err
	}
	if err := encW.Close(); err != nil {
		return err
	}
	if err := tw.WriteHeader(&tar.Header{ Name: "index.zst.aes", Mode: 0600, Size: int64(idxEnc.Len()) }); err != nil {
		return err
	}
	if _, err := tw.Write(idxEnc.Bytes()); err != nil {
		return err
	}

	return nil
}

// classifyPath inspects an os.FileInfo and returns a short file-type code
// ('f' regular, 'd' dir, 'l' symlink, 'p' fifo) and the symlink target.
func classifyPath(path string, fi os.FileInfo) (ft byte, linkname string, err error) {
	mode := fi.Mode()
	if mode.IsRegular() {
		return 'f', "", nil
	}
	if mode.IsDir() {
		return 'd', "", nil
	}
	if mode&os.ModeSymlink != 0 {
		ln, e := os.Readlink(path)
		return 'l', ln, e
	}
	if mode&os.ModeNamedPipe != 0 {
		return 'p', "", nil
	}
	return 0, "", fmt.Errorf("unsupported special file: %s", path)
}

