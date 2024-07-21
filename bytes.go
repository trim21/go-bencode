package bencode

import (
	"errors"
)

type RawBytes []byte

func (b *RawBytes) UnmarshalBencode(bytes []byte) error {
	if b == nil {
		return errors.New("bencode.RawBytes: UnmarshalBencode on nil pointer")
	}
	*b = append((*b)[0:0], bytes...)
	return nil
}

func (b RawBytes) MarshalBencode() ([]byte, error) {
	return b, nil
}

var _ Unmarshaler = (*RawBytes)(nil)
var _ Marshaler = (*RawBytes)(nil)
