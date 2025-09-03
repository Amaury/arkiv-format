package arkivformat

import (
	"archive/tar"
	"fmt"
	"io"
)

// readMagicAndPrefix reads the first two members of the outer tar:
//   1) magic.zst (must decompress to exactly "arkiv001")
//   2) prefix.zst.aes (OpenSSL enc → zstd → 8 random bytes → base64 string)
// It returns the PREFIX_BASE64 string.
func readMagicAndPrefix(tr *tar.Reader, password []byte) (string, error) {
	// 1) Expect and validate magic.zst.
	hdr, err := tr.Next()
	if err != nil {
		return "", err
	}
	if hdr.Name != "magic.zst" {
		return "", fmt.Errorf("expected magic.zst, got %s", hdr.Name)
	}

	// Decompress and verify payload is exactly arkiv001.
	zdecMagic, err := NewZstdDecoder(tr)
	if err != nil {
		return "", err
	}
	payload, err := io.ReadAll(zdecMagic)
	zdecMagic.Close()
	if err != nil {
		return "", err
	}
	if string(payload) != MagicString {
		return "", fmt.Errorf("bad magic")
	}

	// 2) Read prefix.zst.aes and convert to base64 string.
	hdr, err = tr.Next()
	if err != nil {
		return "", err
	}
	if hdr.Name != "prefix.zst.aes" {
		return "", fmt.Errorf("expected prefix.zst.aes, got %s", hdr.Name)
	}

	dr, err := OpenSSLDecryptReader(tr, password)
	if err != nil {
		return "", err
	}
	zdecPrefix, err := NewZstdDecoder(dr)
	if err != nil {
		return "", err
	}
	b8, err := io.ReadAll(zdecPrefix)
	zdecPrefix.Close()
	if err != nil {
		return "", err
	}
	if len(b8) != 8 {
		return "", fmt.Errorf("prefix payload must be 8 bytes, got %d", len(b8))
	}
	return prefixBytesToBase64(b8), nil
}

// scanToParseIndex scans the outer tar stream forward until it finds
// "index.zst.aes", then decrypts and parses it into an Index structure.
func scanToParseIndex(tr *tar.Reader, password []byte) (*Index, error) {
	for {
		hdr, err := tr.Next()
		if err != nil {
			return nil, err
		}
		if hdr.Name == "index.zst.aes" {
			dr, err := OpenSSLDecryptReader(tr, password)
			if err != nil {
				return nil, err
			}
			zdec, err := NewZstdDecoder(dr)
			if err != nil {
				return nil, err
			}
			idx := &Index{}
			s := bufioNewScanner(zdec)
			for s.Scan() {
				line := s.Text()
				if line == "" {
					continue
				}
				raw, hash, perr := parseIndexLine(line)
				if perr != nil {
					zdec.Close()
					return nil, perr
				}
				idx.Entries = append(idx.Entries, IndexEntry{PathRaw: raw, HashData: hash, Quoted: "\"" + raw + "\""})
			}
			if err := s.Err(); err != nil {
				zdec.Close()
				return nil, err
			}
			zdec.Close()
			return idx, nil
		}
	}
}

// bufioNewScanner wraps bufio.NewScanner (split by lines). Separated to
// simplify imports in this utility file.
func bufioNewScanner(r io.Reader) *bufio.Scanner {
	return bufio.NewScanner(r)
}

