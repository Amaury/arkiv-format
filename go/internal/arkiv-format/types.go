package arkivformat

import (
	"crypto/sha512"
)

// Constants for the Arkiv format and environment variables.
const (
	MagicString = "arkiv001"
	EnvPass     = "ARKIV_PASS"
)

// NewSHA512_256 constructs a SHA-512/256 hasher.
// A helper kept for clarity at call sites.
func NewSHA512_256() *sha512.Hash512_256 {
	// Note: type assertion is safe for crypto/sha512 implementation.
	return sha512.New512_256().(*sha512.Hash512_256)
}

