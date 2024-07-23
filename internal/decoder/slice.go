package decoder

import (
	"fmt"
	"reflect"

	"github.com/trim21/go-bencode/internal/errors"
)

type sliceDecoder struct {
	sType             reflect.Type // type of slice
	elemType          reflect.Type // type of element
	isElemPointerType bool
	valueDecoder      Decoder
	structName        string
	fieldName         string
}

func newSliceDecoder(dec Decoder, elemType reflect.Type, structName, fieldName string) *sliceDecoder {
	return &sliceDecoder{
		valueDecoder:      dec,
		elemType:          elemType,
		sType:             reflect.SliceOf(elemType),
		isElemPointerType: elemType.Kind() == reflect.Ptr || elemType.Kind() == reflect.Map,
		structName:        structName,
		fieldName:         fieldName,
	}
}

func (d *sliceDecoder) Decode(ctx *Context, cursor int, depth int64, rv reflect.Value) (int, error) {
	buf := ctx.Buf
	depth++
	if depth > maxDecodeNestingDepth {
		return 0, errors.ErrExceededMaxDepth(buf[cursor], cursor)
	}

	bufSize := len(buf)
	if cursor >= bufSize {
		return 0, errors.DataTooShort(cursor, "list")
	}

	cursor++

	sCap := 8
	index := 0

	s := reflect.New(d.sType).Elem()
	s.Set(reflect.MakeSlice(d.sType, sCap, sCap))

	for {
		if cursor >= bufSize {
			return 0, fmt.Errorf("buffer overflow when decoding dictionary: %d", cursor)
		}

		if buf[cursor] == 'e' {
			if index == sCap-1 { // slice is expensive
				rv.Set(s)
			} else {
				rv.Set(s.Slice(0, index))
			}
			return cursor + 1, nil
		}

		if index == sCap {
			s.Grow(sCap)
			s.Slice(0, sCap)
			sCap = sCap * 2
			s.SetLen(sCap)
		}

		c, err := d.valueDecoder.Decode(ctx, cursor, depth, s.Index(index))
		if err != nil {
			return 0, err
		}

		cursor = c
		index++
	}
}
