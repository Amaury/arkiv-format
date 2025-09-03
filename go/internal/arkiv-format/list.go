package arkivformat

import (
	"archive/tar"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// uidGidToNames tries to resolve uid/gid to local system user/group names.
// Missing entries yield empty strings and callers should fall back to numbers.
func uidGidToNames(uid, gid int) (string, string) {
	var uname string
	var gname string

	u, err := user.LookupId(strconv.Itoa(uid))
	if err == nil && u != nil && u.Username != "" {
		uname = u.Username
	}
	g, err := user.LookupGroupId(strconv.Itoa(gid))
	if err == nil && g != nil && g.Name != "" {
		gname = g.Name
	}
	return uname, gname
}

// ownerString builds a "user:group" string using names when available,
// falling back to numeric ids otherwise.
func ownerString(uid, gid int) string {
	uname, gname := uidGidToNames(uid, gid)
	if uname == "" {
		uname = strconv.Itoa(uid)
	}
	if gname == "" {
		gname = strconv.Itoa(gid)
	}
	return uname + ":" + gname
}

// formatLocalTime formats a UTC timestamp into local time as "YYYY-MM-DD HH:MM".
func formatLocalTime(t time.Time) string {
	return t.In(time.Local).Format("2006-01-02 15:04")
}

// List prints an ls-like listing for entries matching optional prefixes.
// It performs two passes: first to lazily load prefix+index, second to
// iterate tar members and collect meta headers, printing in index order.
func (a *ArchiveReader) List(prefixes []string) error {
	// Ensure we have prefix and index loaded.
	if err := a.ensureLoaded(); err != nil {
		return err
	}

	// Prepare the subset of entries to display.
	wanted := make([]IndexEntry, 0, len(a.index.Entries))
	for _, e := range a.index.Entries {
		if matchesPrefix(e.PathRaw, prefixes) {
			wanted = append(wanted, e)
		}
	}
	if len(wanted) == 0 {
		return nil
	}

	// Build a set of required meta object names.
	required := make(map[string]struct{}, len(wanted))
	for _, e := range wanted {
		hName := computeNameHash(a.prefixB64, e.PathRaw)
		name := filepath.ToSlash(filepath.Join("meta", hName+".tar.zst.aes"))
		required[name] = struct{}{}
	}

	// Map of meta header by object name.
	metas := make(map[string]*tar.Header, len(required))

	// Open the archive for the second pass.
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

	// Walk members to capture the meta we need.
	remaining := len(required)
	for remaining > 0 {
		hdr, err := tr.Next()
		if err != nil {
			return err
		}
		if _, ok := required[hdr.Name]; ok {
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
			metas[hdr.Name] = mh
			remaining--
		}
	}

	// Print output in index order for the selected entries.
	for _, e := range wanted {
		hName := computeNameHash(a.prefixB64, e.PathRaw)
		metaName := filepath.ToSlash(filepath.Join("meta", hName+".tar.zst.aes"))
		mh := metas[metaName]
		if mh == nil {
			return fmt.Errorf("meta chunk not found for %s", e.PathRaw)
		}

		// Pick a single-char type marker.
		var typeCh rune = '-'
		switch mh.Typeflag {
		case tar.TypeDir:
			typeCh = 'd'
		case tar.TypeSymlink:
			typeCh = 'l'
		case tar.TypeFifo:
			typeCh = 'p'
		}

		// Resolve owner and format time in local timezone.
		owner := ownerString(mh.Uid, mh.Gid)
		when := formatLocalTime(mh.ModTime)

		fmt.Printf(
			"%c %04o %s %s %s\n",
			typeCh,
			mh.Mode,
			owner,
			when,
			e.PathRaw,
		)
	}
	return nil
}

// matchesPrefix checks whether a raw path matches any optional prefixes.
// It tries both the raw (escaped) path and a lightly unescaped variant
// (\\" → \" and \\ → \) to ease CLI filtering.
func matchesPrefix(pathRaw string, prefixes []string) bool {
	if len(prefixes) == 0 {
		return true
	}
	unesc := strings.ReplaceAll(pathRaw, "\\", "\\\\")
	unesc = strings.ReplaceAll(unesc, "\"", "\\\"")
	for _, p := range prefixes {
		if strings.HasPrefix(pathRaw, p) {
			return true
		}
		if strings.HasPrefix(unesc, p) {
			return true
		}
	}
	return false
}

