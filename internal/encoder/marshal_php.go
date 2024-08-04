package encoder

import (
	"errors"
	"reflect"
)

var marshalerType = reflect.TypeFor[Marshaler]()

type Marshaler interface {
	MarshalBencode() ([]byte, error)
}

func compileMarshaler(rt reflect.Type) (encoder, error) {
	return func(ctx *Context, b []byte, rv reflect.Value) ([]byte, error) {
		raw, err := rv.Interface().(Marshaler).MarshalBencode()
		if err != nil {
			return nil, err
		}

		if len(raw) == 0 {
			return nil, errors.New("bencode: bencode.Marshaler return empty bytes")
		}

		return append(b, raw...), nil
	}, nil
}
