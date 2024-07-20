package bencode

import (
	"github.com/trim21/go-bencode/internal/encoder"
)

// Marshaler allow users to implement its own encoder.
// **it's return value will not be validated**, please make sure you return valid encoded bytes.
type Marshaler interface {
	MarshalBencode() ([]byte, error)
}

func Marshal(v any) ([]byte, error) {
	return encoder.Marshal(v)
}
