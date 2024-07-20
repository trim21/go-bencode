package decoder

import (
	"reflect"

	"github.com/trim21/go-bencode/internal/errors"
)

type ptrDecoder struct {
	dec        Decoder
	rt         reflect.Type
	structName string
	fieldName  string
}

func newPtrDecoder(dec Decoder, rt reflect.Type, structName, fieldName string) (Decoder, error) {
	if rt.Kind() == reflect.Ptr {
		return nil, &errors.UnsupportedTypeError{
			Type: reflect.PtrTo(rt),
		}
	}
	return &ptrDecoder{
		dec:        dec,
		rt:         rt,
		structName: structName,
		fieldName:  fieldName,
	}, nil
}

func (d *ptrDecoder) Decode(ctx *Context, cursor int, depth int64, rv reflect.Value) (int, error) {
	if rv.IsNil() {
		np := reflect.New(d.rt)
		rv.Set(np)
	}

	c, err := d.dec.Decode(ctx, cursor, depth, rv.Elem())
	if err != nil {
		return 0, err
	}
	cursor = c

	return cursor, nil
}
