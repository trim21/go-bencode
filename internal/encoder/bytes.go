package encoder

import (
	"reflect"
	"strconv"
)

var bytesType = reflect.TypeFor[[]byte]()

func encodeBytesSlice(ctx *Context, b []byte, rv reflect.Value) ([]byte, error) {
	return AppendBytes(b, rv.Bytes()), nil
}

func AppendBytes(b []byte, value []byte) []byte {
	b = strconv.AppendInt(b, int64(len(value)), 10)
	b = append(b, ':')
	return append(b, value...)
}
