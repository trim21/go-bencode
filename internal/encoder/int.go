package encoder

import (
	"math/big"
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

var typeBigIntPtr = reflect.TypeFor[*big.Int]()
var typeBigInt = reflect.TypeFor[big.Int]()

func encodeBigInt(ctx *Context, b []byte, rv reflect.Value) ([]byte, error) {
	v := rv.Interface().(big.Int)

	b = append(b, 'i')
	b = v.Append(b, 10)
	b = append(b, 'e')

	return b, nil
}

func encodeBigIntPtr(ctx *Context, b []byte, rv reflect.Value) ([]byte, error) {
	v := rv.Interface().(*big.Int)

	if v == nil {
		return AppendInt(b, 0), nil
	}

	b = append(b, 'i')
	b = v.Append(b, 10)
	b = append(b, 'e')

	return b, nil
}
