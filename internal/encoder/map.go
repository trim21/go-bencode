package encoder

import (
	"bytes"
	"reflect"
	"slices"
	"strings"
)

// !!! not safe to use in reflect case !!!
func compileMap(rt reflect.Type, seen seenMap) (encoder, error) {
	// for map[int]string, keyType is int, valueType is string
	keyType := rt.Key()
	valueType := rt.Elem()

	var keyEncoder encoder

	var keyCompare func(reflect.Value, reflect.Value) int

	switch {
	case keyType.Kind() == reflect.String:
		keyEncoder = encodeString
		keyCompare = stringKeyCompare
	case keyType.Kind() == reflect.Array && keyType.Elem().Kind() == reflect.Uint8:
		keyEncoder = encodeBytes
		keyCompare = arrayByteKeyCompare
	default:
		return nil, &UnsupportedTypeAsMapKeyError{Type: keyType}
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

		b = append(b, 'd')

		keys := rv.MapKeys()
		slices.SortFunc(keys, keyCompare)

		var err error
		for _, key := range keys {
			b, err = keyEncoder(ctx, b, key)
			if err != nil {
				return b, err
			}

			b, err = valueEncoder(ctx, b, rv.MapIndex(key))
			if err != nil {
				return b, err
			}
		}

		return append(b, 'e'), nil
	}, nil
}

func stringKeyCompare(a reflect.Value, b reflect.Value) int {
	return strings.Compare(a.Interface().(string), b.Interface().(string))
}

func arrayByteKeyCompare(a reflect.Value, b reflect.Value) int {
	return bytes.Compare(a.Bytes(), b.Bytes())
}
