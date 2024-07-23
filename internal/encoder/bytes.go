package encoder

import (
	"reflect"
	"strconv"
)

var bytesType = reflect.TypeFor[[]byte]()

func encodeBytes(ctx *Context, b []byte, rv reflect.Value) ([]byte, error) {
	b = strconv.AppendInt(b, int64(rv.Len()), 10)
	b = append(b, ':')
	return append(b, rv.Bytes()...), nil
}
