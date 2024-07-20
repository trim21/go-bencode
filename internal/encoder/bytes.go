package encoder

import (
	"reflect"
	"strconv"
)

var bytesType = reflect.TypeOf([]byte{})

func encodeBytes(ctx *Ctx, b []byte, rv reflect.Value) ([]byte, error) {
	b = strconv.AppendInt(b, int64(rv.Len()), 10)
	b = append(b, ':')
	return append(b, rv.Bytes()...), nil
}
