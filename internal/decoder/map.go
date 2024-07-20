package decoder

import (
	"fmt"
	"reflect"

	"github.com/trim21/go-bencode/internal/errors"
)

func compileMap(rt reflect.Type, structName, fieldName string, structTypeToDecoder map[reflect.Type]Decoder) (Decoder, error) {
	keyDec, err := compileMapKey(rt.Key(), structName, fieldName, structTypeToDecoder)
	if err != nil {
		return nil, err
	}

	valueDec, err := compile(rt.Elem(), structName, fieldName, structTypeToDecoder)
	if err != nil {
		return nil, err
	}

	return newMapDecoder(rt, rt.Key(), keyDec, rt.Elem(), valueDec, structName, fieldName), nil
}

type mapDecoder struct {
	mapType      reflect.Type
	keyType      reflect.Type
	valueType    reflect.Type
	keyDecoder   Decoder
	valueDecoder Decoder
	structName   string
	fieldName    string
}

func newMapDecoder(mapType reflect.Type, keyType reflect.Type, keyDec Decoder, valueType reflect.Type, valueDec Decoder, structName, fieldName string) *mapDecoder {
	return &mapDecoder{
		mapType:      mapType,
		keyDecoder:   keyDec,
		keyType:      keyType,
		valueType:    valueType,
		valueDecoder: valueDec,
		structName:   structName,
		fieldName:    fieldName,
	}
}

func (d *mapDecoder) Decode(ctx *Context, cursor int, depth int64, rv reflect.Value) (int, error) {
	buf := ctx.Buf

	if buf[cursor] != 'd' {
		return 0, errors.ErrTypeError("dictionary", string(buf[cursor]))
	}

	cursor++

	depth++
	if depth > maxDecodeNestingDepth {
		return 0, errors.ErrExceededMaxDepth(buf[cursor], cursor)
	}

	bufSize := len(buf)
	if bufSize < 2 {
		return 0, errors.ErrExpected("buffer overflow when decoding dictionary", cursor)
	}

	if rv.IsNil() {
		rv.Set(reflect.MakeMapWithSize(d.mapType, 8))
	}

	for {
		if cursor >= bufSize {
			return 0, fmt.Errorf("buffer overflow when decoding dictionary: %d", cursor)
		}

		if buf[cursor] == 'e' {
			cursor++
			return cursor, nil
		}

		k := reflect.New(d.keyType).Elem()
		keyCursor, err := d.keyDecoder.Decode(ctx, cursor, depth, k)
		if err != nil {
			return 0, err
		}
		cursor = keyCursor

		v := reflect.New(d.valueType).Elem()
		valueCursor, err := d.valueDecoder.Decode(ctx, cursor, depth, v)
		if err != nil {
			return 0, err
		}

		rv.SetMapIndex(k, v)
		cursor = valueCursor
	}
}
