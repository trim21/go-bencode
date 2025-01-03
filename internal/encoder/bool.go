package encoder

import (
	"reflect"
)

func encodeBool(ctx *Context, b []byte, rv reflect.Value) ([]byte, error) {
	b = append(b, 'i')

	if rv.Bool() {
		b = append(b, '1')
	} else {
		b = append(b, '0')
	}

	return append(b, 'e'), nil
}
