package encoder

import (
	"errors"
	"fmt"
	"reflect"
)

func compilePtr(rt reflect.Type, seen seenMap) (encoder, error) {
	switch rt.Elem().Kind() {
	case reflect.Ptr:
		return nil, fmt.Errorf("bencode: encoding nested ptr is not supported *%s", rt.Elem().String())
	}

	enc, err := compile(rt.Elem(), seen)
	if err != nil {
		return nil, err
	}

	return deRefNilEncoder(enc), nil
}

var ErrNilPtr = errors.New("bencode: bencode doesn't have a nil type, nil ptr can't be encoded")

func deRefNilEncoder(enc encoder) encoder {
	return func(ctx *Context, b []byte, rv reflect.Value) ([]byte, error) {
		if rv.IsNil() {
			return b, ErrNilPtr
		}

		if ctx.ptrLevel++; ctx.ptrLevel > startDetectingCyclesAfter {
			ptr := rv.UnsafePointer()
			if _, ok := ctx.ptrSeen[ptr]; ok {
				return b, fmt.Errorf("bencode: encountered a cycle via %s", rv.Type())
			}
		}

		var err error
		b, err = enc(ctx, b, rv.Elem())

		ctx.ptrLevel--

		return b, err
	}
}
