package decoder

import (
	"bytes"
	"fmt"
	"reflect"
	"sync"

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
	keyDecoder   Decoder
	valueDecoder Decoder
	structName   string
	fieldName    string
	// temps holds the zero values the key and the value of an entry are decoded
	// into, which SetMapIndex copies into the map and which are zeroed for the
	// next entry, so that decoding an entry allocates neither of them. A pool is
	// per decoder, so the entries of a nested map use other values.
	temps sync.Pool
}

// mapTemps are the zero values of the key and of the value of a map.
type mapTemps struct {
	k, v reflect.Value
}

func newMapDecoder(mapType reflect.Type, keyType reflect.Type, keyDec Decoder, valueType reflect.Type, valueDec Decoder, structName, fieldName string) *mapDecoder {
	d := &mapDecoder{
		mapType:      mapType,
		keyDecoder:   keyDec,
		valueDecoder: valueDec,
		structName:   structName,
		fieldName:    fieldName,
	}
	d.temps.New = func() any {
		return &mapTemps{
			k: reflect.New(keyType).Elem(),
			v: reflect.New(valueType).Elem(),
		}
	}
	return d
}

func (d *mapDecoder) Decode(ctx *Context, cursor int, depth int64, rv reflect.Value) (int, error) {
	buf := ctx.Buf

	bufSize := len(buf)
	if cursor >= bufSize {
		return 0, errors.DataTooShort()
	}

	if buf[cursor] != 'd' {
		return 0, errors.ErrExpecting("dictionary", buf, cursor)
	}

	cursor++

	depth++
	if depth > maxDecodeNestingDepth {
		return 0, errors.ErrExceededMaxDepth(buf[cursor-1], cursor-1)
	}

	if bufSize < 2 {
		return 0, errors.DataTooShort()
	}

	if rv.IsNil() {
		rv.Set(reflect.MakeMapWithSize(d.mapType, 8))
	}

	t := d.temps.Get().(*mapTemps)
	defer d.temps.Put(t)

	k, v := t.k, t.v

	var lastKey []byte

	for {
		if cursor >= bufSize {
			return 0, errors.DataTooShort()
		}

		if buf[cursor] == 'e' {
			cursor++
			return cursor, nil
		}

		currentKey, _, err := readString(buf, cursor)
		if err != nil {
			return 0, err
		}

		k.SetZero()
		keyCursor, err := d.keyDecoder.Decode(ctx, cursor, depth, k)
		if err != nil {
			return 0, err
		}

		if lastKey != nil && !ctx.Relaxed {
			switch bytes.Compare(lastKey, currentKey) {
			case 0:
				return cursor, fmt.Errorf("dictionary conrains duplicated keys %s. index %d", currentKey, cursor)
			case 1:
				return cursor, fmt.Errorf("dictionary conrains unordered keys %s, %s. index %d", lastKey, currentKey, cursor)
			}
		}
		lastKey = currentKey

		cursor = keyCursor
		if cursor >= bufSize {
			return 0, errors.DataTooShort()
		}

		v.SetZero()
		valueCursor, err := d.valueDecoder.Decode(ctx, cursor, depth, v)
		if err != nil {
			return 0, err
		}

		rv.SetMapIndex(k, v)
		cursor = valueCursor
	}
}
