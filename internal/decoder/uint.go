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
	bytes, c, err := decodeIntegerBytes(ctx.Buf, cursor)
	if err != nil {
		return 0, err
	}

	if bytes[0] == '-' {
		return 0, errors.ErrValueOverflow(string(bytes), rv.Type().Kind().String())
	}

	cursor = c

	return d.processBytes(bytes, cursor, rv)
}

func (d *uintDecoder) processBytes(bytes []byte, cursor int, rv reflect.Value) (int, error) {
	u64, err := parseUint64(bytes)
	if err != nil {
		return 0, d.typeError(bytes, cursor)
	}

	if rv.OverflowUint(u64) {
		return 0, errors.ErrValueOverflow(u64, rv.Type().Kind().String())
	}

	rv.SetUint(u64)

	return cursor, nil
}
