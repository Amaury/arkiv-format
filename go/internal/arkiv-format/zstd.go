package arkivformat

import (
	"io"

	"github.com/klauspost/compress/zstd"
)

// NewZstdEncoder creates a zstd encoder writing to w.
func NewZstdEncoder(w io.Writer) (*zstd.Encoder, error) {
	enc, err := zstd.NewWriter(w)
	if err != nil {
		return nil, err
	}
	return enc, nil
}

// NewZstdDecoder creates a zstd decoder reading from r.
func NewZstdDecoder(r io.Reader) (*zstd.Decoder, error) {
	dec, err := zstd.NewReader(r)
	if err != nil {
		return nil, err
	}
	return dec, nil
}

