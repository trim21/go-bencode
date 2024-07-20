package encoder

import (
	"reflect"
	"strconv"
)

func encodeUint(ctx *Context, b []byte, rv reflect.Value) ([]byte, error) {
	b = append(b, 'i')
	b = strconv.AppendUint(b, rv.Uint(), 10)
	return append(b, 'e'), nil
}
