package decoder

import (
	"fmt"
	"reflect"

	"github.com/trim21/go-bencode/internal/errors"
)

type uintDecoder struct {
	rt         reflect.Type
	kind       reflect.Kind
	structName string
	fieldName  string
}

func newUintDecoder(rt reflect.Type, structName, fieldName string) *uintDecoder {
	return &uintDecoder{
		rt:         rt,
		kind:       rt.Kind(),
		structName: structName,
		fieldName:  fieldName,
	}
}

func (d *uintDecoder) typeError(buf []byte, offset int) *errors.UnmarshalTypeError {
	return &errors.UnmarshalTypeError{
		Value:  fmt.Sprintf("number %s", string(buf)),
		Type:   d.rt,
		Offset: offset,
	}
}

func (d *uintDecoder) Decode(ctx *Context, cursor int, depth int64, rv reflect.Value) (int, error) {
	bytes, num, neg, overflow, c, err := decodeIntegerBytes(ctx.Buf, cursor)
	if err != nil {
		return 0, err
	}

	if neg {
		return 0, errors.ErrValueOverflow(string(bytes), rv.Type().Kind().String())
	}

	if overflow {
		return 0, d.typeError(bytes, c)
	}

	if rv.OverflowUint(num) {
		return 0, errors.ErrValueOverflow(num, rv.Type().Kind().String())
	}

	rv.SetUint(num)

	return c, nil
}
