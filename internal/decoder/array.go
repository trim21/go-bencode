package decoder

import (
	"fmt"
	"reflect"

	"github.com/trim21/go-bencode/internal/errors"
)

type arrayDecoder struct {
	elemType     reflect.Type
	valueDecoder Decoder
	aType        reflect.Type
	alen         int
	structName   string
	fieldName    string
	zeroValue    reflect.Value
}

func newArrayDecoder(dec Decoder, elemType reflect.Type, alen int, structName, fieldName string) *arrayDecoder {
	zeroValue := reflect.Zero(elemType)
	return &arrayDecoder{
		valueDecoder: dec,
		elemType:     elemType,
		aType:        reflect.ArrayOf(alen, elemType),
		alen:         alen,
		structName:   structName,
		fieldName:    fieldName,
		zeroValue:    zeroValue,
	}
}

func (d *arrayDecoder) Decode(ctx *Context, cursor int, depth int64, rv reflect.Value) (int, error) {
	buf := ctx.Buf
	bufSize := len(buf)

	depth++
	if depth > maxDecodeNestingDepth {
		return 0, errors.ErrExceededMaxDepth(buf[cursor], cursor)
	}

	if cursor >= bufSize {
		return 0, errors.DataTooShort(cursor, "list")
	}

	if buf[cursor] != 'l' {
		return 0, errors.ErrTypeMismatch(d.aType.String(), string(buf[cursor]))
	}

	cursor++

	index := 0

	for {
		if cursor >= bufSize {
			return 0, fmt.Errorf("buffer overflow when decoding dictionary: %d", cursor)
		}

		if buf[cursor] == 'e' {
			if index != d.alen-1 {
				return 0, fmt.Errorf("bencode: failed to decode list into array, list length %d. array length %d", index+1, d.alen)
			}
			return cursor + 1, nil
		}

		if index >= d.alen {
			return 0, fmt.Errorf("array overflow when decoding list: index %d", cursor)
		}

		c, err := d.valueDecoder.Decode(ctx, cursor, depth, rv.Index(index))
		if err != nil {
			return 0, err
		}

		cursor = c
		index++
	}
}
