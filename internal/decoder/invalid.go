package decoder

import (
	"reflect"

	"github.com/trim21/go-bencode/internal/errors"
)

type invalidDecoder struct {
	rt         reflect.Type
	kind       reflect.Kind
	structName string
	fieldName  string
}

func newInvalidDecoder(rt reflect.Type, structName, fieldName string) *invalidDecoder {
	return &invalidDecoder{
		rt:         rt,
		kind:       rt.Kind(),
		structName: structName,
		fieldName:  fieldName,
	}
}

func (d *invalidDecoder) Decode(ctx *Context, cursor int, depth int64, rv reflect.Value) (int, error) {
	return 0, &errors.UnmarshalTypeError{
		Value:  "object",
		Type:   d.rt,
		Offset: cursor,
		Struct: d.structName,
		Field:  d.fieldName,
	}
}
