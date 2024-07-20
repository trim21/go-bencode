package bencode

import (
	"errors"
)

type Bytes []byte

func (b *Bytes) UnmarshalBencode(bytes []byte) error {
	if b == nil {
		return errors.New("bencode.Bytes: UnmarshalBencode on nil pointer")
	}
	*b = append((*b)[0:0], bytes...)
	return nil
}

func (b Bytes) MarshalBencode() ([]byte, error) {
	return b, nil
}

var _ Unmarshaler = (*Bytes)(nil)
var _ Marshaler = (*Bytes)(nil)
