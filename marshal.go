package bencode

import (
	"github.com/trim21/go-bencode/internal/encoder"
)

// Marshaler allow users to implement its own encoder.
type Marshaler interface {
	MarshalBencode() ([]byte, error)
}

// IsZeroValue add support type implements Marshaler and omitempty
//
//	var s struct {
//		Field T `bencode:"field,omitempty"`
//	}
//
// if `T` implements  [Marshaler], it can implement [IsZeroValue] so bencode know if it's
type IsZeroValue interface {
	// IsZeroBencodeValue enable support for omitempty feature.
	// if it's being used as struct field, and it returns true, this field will be skipped.
	IsZeroBencodeValue() bool
}

func Marshal(v any) ([]byte, error) {
	return encoder.Marshal(v)
}
