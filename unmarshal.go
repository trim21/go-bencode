package bencode

import (
	"errors"

	"github.com/trim21/go-bencode/internal/decoder"
)

type Unmarshaler interface {
	UnmarshalBencode([]byte) error
}

func Unmarshal(data []byte, v any) error {
	if len(data) == 0 {
		return errors.New("empty data")
	}

	return decoder.Unmarshal(data, v)
}

// UnmarshalRelaxed is like Unmarshal but with relaxed parsing rules:
// - Dictionary keys are not required to be sorted
// - Duplicate dictionary keys are allowed (last value wins)
func UnmarshalRelaxed(data []byte, v any) error {
	if len(data) == 0 {
		return errors.New("empty data")
	}

	return decoder.UnmarshalRelaxed(data, v)
}
