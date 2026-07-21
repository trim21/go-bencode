//go:build 1.23

package encoder

import (
	"reflect"
	"strconv"
)

func compileBytesArray(rt reflect.Type) (encoder, error) {
	var head []byte

	size := rt.Len()

	head = strconv.AppendInt(head, int64(rt.Len()), 10)
	head = append(head, ':')

	return func(ctx *Context, b []byte, rv reflect.Value) ([]byte, error) {
		b = append(b, head...)

		if rv.CanAddr() {
			return append(b, rv.Bytes()...), nil
		}

		for i := range size {
			b = append(b, byte(rv.Index(i).Uint()))
		}

		return b, nil
	}, nil
}

// byteArrayToBytes returns a []byte view of a [N]byte array.
// For addressable values it's zero-copy via Bytes(); for unaddressable
// (e.g. from MapKeys()) it copies into a new slice.
func byteArrayToBytes(rv reflect.Value) []byte {
	if rv.CanAddr() {
		return rv.Bytes()
	}
	b := make([]byte, rv.Len())
	for i := range b {
		b[i] = byte(rv.Index(i).Uint())
	}
	return b
}
