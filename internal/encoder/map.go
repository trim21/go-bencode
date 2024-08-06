package encoder

import (
	"bytes"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

const startDetectingCyclesAfter = 1000

// !!! not safe to use in reflect case !!!
func compileMap(rt reflect.Type, seen seenMap) (encoder, error) {
	// for map[int]string, keyType is int, valueType is string
	keyType := rt.Key()
	valueType := rt.Elem()

	var keyEncoder encoder
	var err error

	var keyCompare func(reflect.Value, reflect.Value) int

	switch {
	case keyType.Kind() == reflect.String:
		keyEncoder = encodeString
		keyCompare = stringKeyCompare
	case keyType.Kind() == reflect.Array && keyType.Elem().Kind() == reflect.Uint8:
		keyEncoder, err = compileBytesArray(keyType)
		keyCompare = arrayByteKeyCompare
	default:
		return nil, &UnsupportedTypeAsMapKeyError{Type: keyType}
	}

	if err != nil {
		return nil, err
	}

	valueEncoder, err := compile(valueType, seen)
	if err != nil {
		return nil, err
	}

	return func(ctx *Context, b []byte, rv reflect.Value) ([]byte, error) {
		if rv.IsNil() {
			return appendEmptyMap(b), nil
		}

		size := rv.Len()
		if size == 0 {
			return appendEmptyMap(b), nil
		}

		if ctx.ptrLevel++; ctx.ptrLevel > startDetectingCyclesAfter {
			ptr := rv.UnsafePointer()
			if _, ok := ctx.ptrSeen[ptr]; ok {
				return b, fmt.Errorf("bencode: encountered a cycle via %s", rv.Type())
			}
		}

		b = append(b, 'd')

		keys := rv.MapKeys()
		slices.SortFunc(keys, keyCompare)

		var err error
		for _, key := range keys {
			b, err = keyEncoder(ctx, b, key)
			if err != nil {
				return b, fmt.Errorf("encountered a cycle via %s", rv.Type())
			}

			b, err = valueEncoder(ctx, b, rv.MapIndex(key))
			if err != nil {
				return b, err
			}
		}

		ctx.ptrLevel--

		return append(b, 'e'), nil
	}, nil
}

func stringKeyCompare(a reflect.Value, b reflect.Value) int {
	return strings.Compare(a.String(), b.String())
}

func arrayByteKeyCompare(a reflect.Value, b reflect.Value) int {
	return bytes.Compare(a.Bytes(), b.Bytes())
}

func appendEmptyMap(b []byte) []byte {
	return append(b, "de"...)
}
