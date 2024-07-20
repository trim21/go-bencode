package decoder

import (
	"reflect"

	"github.com/trim21/go-bencode/internal/errors"
)

type boolDecoder struct {
	structName string
	fieldName  string
}

func newBoolDecoder(structName, fieldName string) *boolDecoder {
	return &boolDecoder{structName: structName, fieldName: fieldName}
}

func (d *boolDecoder) Decode(ctx *Context, cursor int, depth int64, rv reflect.Value) (int, error) {
	buf := ctx.Buf
	if cursor > len(buf)-3 {
		return 0, errors.ErrSyntax("invalid int", cursor)
	}

	switch buf[cursor] {
	case 'i':
		// i0e;
		// i1e;

		cursor++
		switch buf[cursor] {
		case '0':
			rv.SetBool(false)
		case '1':
			rv.SetBool(true)
		default:
			return 0, errors.ErrInvalidCharacter(buf[cursor], "bool value", cursor)
		}

		cursor++
		if buf[cursor] != 'e' {
			return 0, errors.ErrUnexpectedEnd("'e' end bool value", cursor)
		}

		cursor++
		return cursor, nil
	}

	return 0, errors.ErrUnexpectedEnd("bool", cursor)
}
