package decoder

import (
	"reflect"
)

type stringDecoder struct {
	structName string
	fieldName  string
}

func newStringDecoder(structName, fieldName string) *stringDecoder {
	return &stringDecoder{
		structName: structName,
		fieldName:  fieldName,
	}
}

func (d *stringDecoder) Decode(ctx *Context, cursor int, depth int64, rv reflect.Value) (int, error) {
	bytes, c, err := readString(ctx.Buf, cursor)
	if err != nil {
		return 0, err
	}
	if len(bytes) != 0 {
		rv.SetString(string(bytes))
	}
	return c, nil
}
