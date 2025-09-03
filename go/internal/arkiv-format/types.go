package arkivformat

import (
	"crypto/sha512"
	"hash"
)

// Constants for the Arkiv format and environment variables.
const (
	MagicString = "arkiv001"
	EnvPass     = "ARKIV_PASS"
)

// NewSHA512_256 constructs a SHA-512/256 hasher and return it as a
// generic hash.Hash. Returning the interface avoids relying on any
// unexported concrete type from crypto/sha512.
func NewSHA512_256() hash.Hash {
	// Note: type assertion is safe for crypto/sha512 implementation.
	return sha512.New512_256()
}

