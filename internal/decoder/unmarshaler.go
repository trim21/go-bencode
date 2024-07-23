package decoder

import (
	"reflect"

	"github.com/trim21/go-bencode/internal/errors"
)

type Unmarshaler interface {
	UnmarshalBencode([]byte) error
}

var (
	unmarshalerType = reflect.TypeFor[Unmarshaler]()
)

type unmarshalerDecoder struct {
	rt         reflect.Type
	structName string
	fieldName  string
}

func newUnmarshalerDecoder(rt reflect.Type, structName, fieldName string) *unmarshalerDecoder {
	return &unmarshalerDecoder{
		rt:         rt,
		structName: structName,
		fieldName:  fieldName,
	}
}

func (d *unmarshalerDecoder) annotateError(cursor int, err error) {
	switch e := err.(type) {
	case *errors.UnmarshalTypeError:
		e.Struct = d.structName
		e.Field = d.fieldName
	case *errors.SyntaxError:
		e.Offset = cursor
	}
}

func (d *unmarshalerDecoder) Decode(ctx *Context, cursor int, depth int64, rv reflect.Value) (int, error) {
	buf := ctx.Buf
	start := cursor
	end, err := skipValue(buf, cursor, depth)
	if err != nil {
		return 0, err
	}
	src := buf[start:end]

	v := reflect.New(d.rt.Elem())

	if err := v.Interface().(Unmarshaler).UnmarshalBencode(src); err != nil {
		d.annotateError(cursor, err)
		return 0, err
	}

	rv.Set(v.Elem())

	return end, nil
}
