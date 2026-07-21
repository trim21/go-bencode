//go:build !1.23

package encoder

import (
	"reflect"
	"strconv"
	"unsafe"
)

func compileBytesArray(rt reflect.Type) (encoder, error) {
	var head []byte

	head = strconv.AppendInt(head, int64(rt.Len()), 10)
	head = append(head, ':')

	return func(ctx *Context, b []byte, rv reflect.Value) ([]byte, error) {
		b = append(b, head...)
		var buf []byte

		if rv.CanAddr() {
			buf = rv.Bytes()
		} else {
			buf = unsafeBytesFromArray(rv)
		}

		return append(b, buf...), nil
	}, nil
}

type eface struct {
	_    uintptr
	data unsafe.Pointer
}

func unsafeBytesFromArray(rv reflect.Value) []byte {
	v := rv.Interface()
	ef := (*eface)(unsafe.Pointer(&v))
	return unsafe.Slice((*byte)(ef.data), rv.Len())
}

// byteArrayToBytes returns a []byte view of a [N]byte array without copying.
// Handles the unaddressable case (e.g. from MapKeys()) via unsafe.
func byteArrayToBytes(rv reflect.Value) []byte {
	if rv.CanAddr() {
		return rv.Bytes()
	}
	return unsafeBytesFromArray(rv)
}
