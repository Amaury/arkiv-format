package arkivformat

import (
	"archive/tar"
	"os"
)

// ArchiveReader represents a read session for an Arkiv archive.
// It encapsulates the archive path, password, and lazily loaded state
// like the PREFIX_BASE64 and the parsed textual index.
type ArchiveReader struct {
	path      string
	password  []byte
	prefixB64 string
	index     *Index
}

// NewArchiveReader creates a new reader session for the given archive path
// and password. It does not immediately read the archive; loading is lazy.
func NewArchiveReader(path string, password []byte) *ArchiveReader {
	return &ArchiveReader{path: path, password: password}
}

// ensureLoaded lazily initializes prefixB64 and the textual index by
// reading the magic, prefix, and scanning forward to index.zst.aes.
func (a *ArchiveReader) ensureLoaded() error {
	// If already loaded, nothing to do.
	if a.index != nil && a.prefixB64 != "" {
		return nil
	}

	// Open the archive for the first pass.
	f, err := os.Open(a.path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Create a tar reader and validate the magic and prefix members.
	tr := tar.NewReader(f)
	prefix, err := readMagicAndPrefix(tr, a.password)
	if err != nil {
		return err
	}

	// Scan forward until index.zst.aes and parse it.
	idx, err := scanToParseIndex(tr, a.password)
	if err != nil {
		return err
	}

	// Cache for subsequent operations.
	a.prefixB64 = prefix
	a.index = idx
	return nil
}

// Close attempts to securely wipe the password bytes. It does not close
// any files (they are managed per method).
func (a *ArchiveReader) Close() {
	if a.password != nil {
		for i := range a.password {
			a.password[i] = 0
		}
	}
}

// ArchiveWriter represents a write session for creating Arkiv archives.
// It encapsulates the destination path and the password used for encryption.
type ArchiveWriter struct {
	path     string
	password []byte
}

// NewArchiveWriter constructs a writer session for a target archive path
// and password. The archive is created when Create(...) is called.
func NewArchiveWriter(path string, password []byte) *ArchiveWriter {
	return &ArchiveWriter{path: path, password: password}
}

// Close attempts to securely wipe the password bytes.
func (w *ArchiveWriter) Close() {
	if w.password != nil {
		for i := range w.password {
			w.password[i] = 0
		}
	}
}

