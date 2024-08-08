package encoder

import (
	"reflect"
)

func compileSlice(rt reflect.Type, seen seenMap) (encoder, error) {
	var enc encoder
	var err error

	if rt == bytesType {
		return encodeBytesSlice, nil
	}

	enc, err = compile(rt.Elem(), seen)
	if err != nil {
		return nil, err
	}

	return func(ctx *Context, b []byte, rv reflect.Value) ([]byte, error) {
		if rv.IsNil() {
			return appendEmptyList(b), nil
		}

		b = append(b, 'l')

		length := rv.Len()

		var err error // create a new error value, so shadow compiler's error
		for i := 0; i < length; i++ {
			b, err = enc(ctx, b, rv.Index(i))
			if err != nil {
				return b, err
			}
		}
		return append(b, 'e'), nil
	}, nil
}

func appendEmptyList(b []byte) []byte {
	return append(b, "le"...)
}
