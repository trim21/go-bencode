package encoder

import (
	"errors"
	"fmt"
	"reflect"
)

func compilePtr(rt reflect.Type, seen seenMap) (encoder, error) {
	switch rt.Elem().Kind() {
	case reflect.Pointer:
		return nil, fmt.Errorf("bencode: encoding nested ptr is not supported *%s", rt.Elem().String())
	}

	inner, err := compile(rt.Elem(), seen)
	if err != nil {
		return nil, err
	}

	elemStruct := rt.Elem().Kind() == reflect.Struct

	return func(ctx *Context, b []byte, rv reflect.Value) ([]byte, error) {
		if rv.IsNil() {
			return b, ErrNilPtr
		}

		if elemStruct {
			if ctx.depth++; ctx.depth > startDetectingCyclesAfter {
				ptr := rv.UnsafePointer()
				if _, ok := ctx.ptrSeen[ptr]; ok {
					return b, fmt.Errorf("bencode: encountered a cycle via %s", rv.Type())
				}
				ctx.ptrSeen[ptr] = empty{}
				defer delete(ctx.ptrSeen, ptr)
			}
		}

		var err error
		b, err = inner(ctx, b, rv.Elem())

		if elemStruct {
			ctx.depth--
		}

		return b, err
	}, nil
}

var ErrNilPtr = errors.New("bencode: bencode doesn't have a nil type, nil ptr can't be encoded")
