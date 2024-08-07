package decoder

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"

	"github.com/trim21/go-bencode/internal/errors"
)

type interfaceDecoder struct {
	rt         reflect.Type
	structName string
	fieldName  string
}

func newEmptyInterfaceDecoder(rt reflect.Type, structName, fieldName string) *interfaceDecoder {
	return &interfaceDecoder{
		rt:         rt,
		structName: structName,
		fieldName:  fieldName,
	}
}

func newInterfaceDecoder(rt reflect.Type, structName, fieldName string) *interfaceDecoder {
	return newEmptyInterfaceDecoder(rt, structName, fieldName)
}

func (d *interfaceDecoder) Decode(ctx *Context, cursor int, depth int64, rv reflect.Value) (int, error) {
	buf := ctx.Buf
	if cursor >= len(buf) {
		return 0, errors.ErrSyntax("input too short when decoding any", cursor)
	}

	v, end, err := d.decodeAny(ctx, cursor)
	if err != nil {
		return 0, err
	}

	rv.Set(reflect.ValueOf(v))

	return end, nil
}

func (d *interfaceDecoder) decodeAny(ctx *Context, cursor int) (any, int, error) {
	buf := ctx.Buf
	if cursor >= len(buf) {
		return nil, 0, errors.ErrSyntax("input too short when decoding any", cursor)
	}

	switch buf[cursor] {
	case 'd':
		return d.decodeDict(ctx, cursor)
	case 'l':
		return d.decodeList(ctx, cursor)
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		b, end, err := readString(buf, cursor)
		if err != nil {
			return nil, 0, err
		}
		return string(b), end, nil
	case 'i':
		v, end, err := decodeIntegerBytes(buf, cursor)
		if err != nil {
			return nil, 0, err
		}
		i, err := strconv.ParseInt(string(v), 10, 64)
		return i, end, err
	}

	return nil, cursor, errors.ErrInvalidBeginningOfValue(buf[cursor], cursor)
}

func (d *interfaceDecoder) decodeList(ctx *Context, cursor int) ([]any, int, error) {
	buf := ctx.Buf

	bufSize := len(buf)
	if cursor >= bufSize {
		return nil, 0, errors.DataTooShort()
	}

	cursor++

	if bufSize < 2 {
		return nil, 0, errors.DataTooShort()
	}

	var r = make([]any, 0, 8)

	for {
		if cursor >= bufSize {
			return nil, 0, errors.DataTooShort()
		}

		if buf[cursor] == 'e' {
			cursor++
			return r, cursor, nil
		}

		v, end, err := d.decodeAny(ctx, cursor)
		if err != nil {
			return nil, 0, err
		}

		r = append(r, v)

		cursor = end
	}
}

func (d *interfaceDecoder) decodeDict(ctx *Context, cursor int) (map[string]any, int, error) {
	buf := ctx.Buf

	bufSize := len(buf)
	if cursor >= bufSize {
		return nil, 0, errors.DataTooShort()
	}

	cursor++

	if bufSize < 2 {
		return nil, 0, errors.DataTooShort()
	}

	var m = make(map[string]any, 8)

	var lastKey []byte

	for {
		if cursor >= bufSize {
			return nil, 0, errors.DataTooShort()
		}

		if buf[cursor] == 'e' {
			cursor++
			return m, cursor, nil
		}

		rawKey, keyCursor, err := readString(buf, cursor)
		if err != nil {
			return nil, 0, err
		}

		if lastKey != nil {
			switch bytes.Compare(lastKey, rawKey) {
			case 0:
				return nil, cursor, fmt.Errorf("dictionary conrains duplicated keys %s. index %d", rawKey, cursor)
			case 1:
				return nil, cursor, fmt.Errorf("dictionary conrains unordered keys %s, %s. index %d", lastKey, rawKey, cursor)
			}
		}

		lastKey = rawKey
		cursor = keyCursor

		v, end, err := d.decodeAny(ctx, cursor)
		if err != nil {
			return nil, 0, err
		}

		m[string(rawKey)] = v

		cursor = end
	}
}
