package arkivformat

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// IndexEntry represents one logical line in the textual index.
// PathRaw holds the exact substring between quotes as stored (with any
// backslashes kept as-is). HashData holds the lowercase hex hash for
// regular files (empty otherwise). Quoted contains the `"<escaped>"` form.
type IndexEntry struct {
	PathRaw  string
	HashData string
	Quoted   string
}

// Index holds all entries of the archive index and provides serialization
// into the canonical byte-wise sorted format used in index.zst.aes.
type Index struct {
	Entries []IndexEntry
}

// escapeForIndex applies the exact escaping rules used by the shell:
//   - Replace backslash (\) with \\ (double backslash)
//   - Replace double quote (") with \\" (backslash-quote)
// Then wrap with surrounding double quotes. It returns both the quoted
// form and the raw-between-quotes content.
func escapeForIndex(path string) (quoted string, rawBetween string) {
	// Escape backslashes first, then quotes.
	r := strings.ReplaceAll(path, "\", "\\")
	r = strings.ReplaceAll(r, "\"", "\\"")

	// Build final quoted representation.
	quoted = "\"" + r + "\""
	rawBetween = r
	return
}

// parseIndexLine parses one line of the index of the form:
//   "PATH"
// or
//   "PATH"=HASH
// There must be no spaces. It returns the raw PATH substring (as-is) and an
// optional hash. It never unescapes PATH.
func parseIndexLine(line string) (raw string, hash string, err error) {
	// Must start with a double quote.
	if !strings.HasPrefix(line, "\"") {
		return "", "", fmt.Errorf("bad index line: %q", line)
	}

	// Find the closing double quote.
	i := strings.IndexByte(line[1:], '"')
	if i < 0 {
		return "", "", fmt.Errorf("unterminated path: %q", line)
	}
	i++ // adjust index since we searched from line[1:]

	// Extract raw substring without quotes.
	raw = line[1:i]

	// If nothing follows, there is no hash.
	if len(line) == i+1 {
		return raw, "", nil
	}

	// Otherwise expect '=' then the hex hash.
	if i+1 >= len(line) || line[i+1] != '=' {
		return "", "", fmt.Errorf("bad index sep: %q", line)
	}
	hash = line[i+2:]
	return raw, hash, nil
}

// Serialize returns the canonical textual content of the index:
//   - Deduplicate exact lines byte-wise
//   - Sort them using byte ordering (LC_ALL=C)
//   - Join with '\n' (no trailing newline)
func (idx *Index) Serialize() []byte {
	seen := make(map[string]struct{}, len(idx.Entries))
	lines := make([][]byte, 0, len(idx.Entries))

	for _, e := range idx.Entries {
		line := e.Quoted
		if e.HashData != "" {
			line += "=" + strings.ToLower(e.HashData)
		}
		if _, ok := seen[line]; ok {
			continue
		}
		seen[line] = struct{}{}
		lines = append(lines, []byte(line))
	}

	sort.Slice(lines, func(i, j int) bool {
		return bytes.Compare(lines[i], lines[j]) < 0
	})
	return bytes.Join(lines, []byte{'\n'})
}

// computeNameHash returns SHA-512/256( PREFIX_BASE64 || PATH_BYTES ) where
// PATH_BYTES is the exact substring between quotes (no unescaping).
func computeNameHash(prefixB64 string, pathRaw string) string {
	h := sha512.New512_256()
	_, _ = h.Write([]byte(prefixB64))
	_, _ = h.Write([]byte(pathRaw))
	return hex.EncodeToString(h.Sum(nil))
}

// prefixBytesToBase64 encodes the 8 random prefix bytes to a single-line
// Base64 string without trailing newline.
func prefixBytesToBase64(b8 []byte) string {
	return base64.StdEncoding.EncodeToString(b8)
}

