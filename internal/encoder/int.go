package encoder

import (
	"reflect"
	"strconv"
)

func encodeInt(ctx *Context, b []byte, rv reflect.Value) ([]byte, error) {
	return AppendInt(b, rv.Int()), nil
}

func AppendInt(b []byte, v int64) []byte {
	b = append(b, 'i')
	b = strconv.AppendInt(b, v, 10)
	return append(b, 'e')
}
