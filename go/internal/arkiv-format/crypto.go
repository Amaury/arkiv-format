package arkivformat

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

// OpenSSL-compatible constants.
const (
	opensslHeader = "Salted__"
	pbkdf2Iter    = 10000
	keyLen        = 32
	ivLen         = 16
)

// OpenSSLEncryptWriter returns a WriteCloser that emits the OpenSSL header
// ("Salted__" + 8-byte salt) followed by AES-256-CBC encrypted data with
// PKCS#7 padding. The plaintext written to the returned writer will be
// encrypted on Close().
func OpenSSLEncryptWriter(w io.Writer, password []byte) (io.WriteCloser, error) {
	// Generate an 8-byte salt.
	salt := make([]byte, 8)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}

	// Emit the OpenSSL header and salt.
	if _, err := w.Write([]byte(opensslHeader)); err != nil {
		return nil, err
	}
	if _, err := w.Write(salt); err != nil {
		return nil, err
	}

	// Derive key and IV using PBKDF2-HMAC-SHA256 with 10000 iterations.
	keyiv := pbkdf2.Key(password, salt, pbkdf2Iter, keyLen+ivLen, sha256.New)
	block, err := aes.NewCipher(keyiv[:keyLen])
	if err != nil {
		return nil, err
	}
	iv := keyiv[keyLen:]
	mode := cipher.NewCBCEncrypter(block, iv)

	// Wrap in a CBC+PKCS7 writer.
	return newCBCPKCS7Writer(w, mode), nil
}

// OpenSSLDecryptReader consumes an OpenSSL-compatible stream and returns
// a reader that yields the decrypted plaintext.
func OpenSSLDecryptReader(r io.Reader, password []byte) (io.Reader, error) {
	// Read and validate the header.
	head := make([]byte, 8)
	if _, err := io.ReadFull(r, head); err != nil {
		return nil, err
	}
	if string(head) != opensslHeader {
		return nil, errors.New("invalid OpenSSL header")
	}

	// Read the 8-byte salt and derive key/iv.
	salt := make([]byte, 8)
	if _, err := io.ReadFull(r, salt); err != nil {
		return nil, err
	}
	keyiv := pbkdf2.Key(password, salt, pbkdf2Iter, keyLen+ivLen, sha256.New)
	block, err := aes.NewCipher(keyiv[:keyLen])
	if err != nil {
		return nil, err
	}
	iv := keyiv[keyLen:]
	mode := cipher.NewCBCDecrypter(block, iv)

	// Wrap in a CBC+PKCS7 reader.
	return newCBCPKCS7Reader(r, mode), nil
}

// cbcPKCS7Writer buffers plaintext, encrypts full blocks as they become
// available, and on Close() applies PKCS#7 padding and flushes the final
// encrypted blocks.
type cbcPKCS7Writer struct {
	w    io.Writer
	mode cipher.BlockMode
	buf  []byte
}

// newCBCPKCS7Writer constructs the streaming writer.
func newCBCPKCS7Writer(w io.Writer, mode cipher.BlockMode) io.WriteCloser {
	return &cbcPKCS7Writer{w: w, mode: mode}
}

// Write buffers p and encrypts any full blocks to the underlying writer.
func (c *cbcPKCS7Writer) Write(p []byte) (int, error) {
	// Append to internal buffer.
	c.buf = append(c.buf, p...)

	// Determine how many bytes form complete blocks.
	blockSize := c.mode.BlockSize()
	n := len(c.buf) / blockSize * blockSize
	if n == 0 {
		return len(p), nil
	}

	// Encrypt full blocks and write them.
	toEnc := c.buf[:n]
	enc := make([]byte, len(toEnc))
	c.mode.CryptBlocks(enc, toEnc)
	if _, err := c.w.Write(enc); err != nil {
		return 0, err
	}

	// Keep the remainder (less than one block) for the next write or Close.
	c.buf = c.buf[n:]
	return len(p), nil
}

// Close applies PKCS#7 padding to the remaining bytes and flushes them.
func (c *cbcPKCS7Writer) Close() error {
	blockSize := c.mode.BlockSize()
	padLen := blockSize - (len(c.buf) % blockSize)
	if padLen == 0 {
		padLen = blockSize
	}

	// Append padding bytes (each byte equals pad length).
	for i := 0; i < padLen; i++ {
		c.buf = append(c.buf, byte(padLen))
	}

	// Encrypt and flush the final padded blocks.
	enc := make([]byte, len(c.buf))
	c.mode.CryptBlocks(enc, c.buf)
	_, err := c.w.Write(enc)
	return err
}

// cbcPKCS7Reader decrypts incoming ciphertext blocks and removes PKCS#7
// padding on the final read. It maintains internal buffers to return
// plaintext in any slice sizes requested by the caller.
type cbcPKCS7Reader struct {
	r    io.Reader
	mode cipher.BlockMode
	buf  []byte
	out  []byte
	fin  bool
}

// newCBCPKCS7Reader constructs the streaming reader.
func newCBCPKCS7Reader(r io.Reader, mode cipher.BlockMode) io.Reader {
	return &cbcPKCS7Reader{r: r, mode: mode}
}

// Read decrypts full blocks when available, defers the final block until
// the underlying reader returns EOF, then validates and strips padding.
func (c *cbcPKCS7Reader) Read(p []byte) (int, error) {
	// If we have leftover plaintext from a previous call, serve it first.
	if len(c.out) > 0 {
		n := copy(p, c.out)
		c.out = c.out[n:]
		return n, nil
	}

	// If we've already finished, propagate EOF.
	if c.fin {
		return 0, io.EOF
	}

	// Read more ciphertext from the underlying reader.
	buf := make([]byte, 4096)
	nr, err := c.r.Read(buf)
	if err != nil && err != io.EOF {
		return 0, err
	}
	c.buf = append(c.buf, buf[:nr]...)

	// Only decrypt up to the last full block; keep any tail for next time.
	blockSize := c.mode.BlockSize()
	n := len(c.buf) / blockSize * blockSize
	if err == io.EOF {
		// Mark finalization so we can remove padding after decrypting.
		c.fin = true
	}
	if n == 0 {
		// Not enough to decrypt a whole block yet.
		if c.fin {
			// EOF but no full block is an error for CBC.
			return 0, io.ErrUnexpectedEOF
		}
		return 0, nil
	}

	// Decrypt the available full blocks.
	dec := make([]byte, n)
	c.mode.CryptBlocks(dec, c.buf[:n])
	c.buf = c.buf[n:]

	// On finalization, validate and remove PKCS#7 padding bytes.
	if c.fin {
		if len(dec) < blockSize {
			return 0, errors.New("invalid padding: short final block")
		}
		padLen := int(dec[len(dec)-1])
		if padLen == 0 || padLen > blockSize {
			return 0, errors.New("invalid padding: range")
		}
		for i := 0; i < padLen; i++ {
			if dec[len(dec)-1-i] != byte(padLen) {
				return 0, errors.New("invalid padding: content")
			}
		}
		dec = dec[:len(dec)-padLen]
	}

	// Serve from decrypted bytes; keep leftovers for next call.
	nw := copy(p, dec)
	if nw < len(dec) {
		c.out = dec[nw:]
	}
	if c.fin && len(dec) == 0 && len(c.out) == 0 {
		return 0, io.EOF
	}
	return nw, nil
}

