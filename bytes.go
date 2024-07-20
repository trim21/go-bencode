package bencode

import (
	"errors"
)

type RawMessage []byte

func (b *RawMessage) UnmarshalBencode(bytes []byte) error {
	if b == nil {
		return errors.New("bencode.RawMessage: UnmarshalBencode on nil pointer")
	}
	*b = append((*b)[0:0], bytes...)
	return nil
}

func (b RawMessage) MarshalBencode() ([]byte, error) {
	return b, nil
}

var _ Unmarshaler = (*RawMessage)(nil)
var _ Marshaler = (*RawMessage)(nil)
