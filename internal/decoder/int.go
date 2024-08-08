package decoder

import (
	"bytes"
	"fmt"
	"math/big"
	"reflect"
	"strconv"

	"github.com/trim21/go-bencode/internal/errors"
)

type intDecoder struct {
	rt         reflect.Type
	kind       reflect.Kind
	structName string
	fieldName  string
}

func newIntDecoder(rt reflect.Type, structName, fieldName string) *intDecoder {
	return &intDecoder{
		rt:         rt,
		kind:       rt.Kind(),
		structName: structName,
		fieldName:  fieldName,
	}
}

func decodeIntegerBytes(buf []byte, cursor int) ([]byte, int, error) {
	if buf[cursor] != 'i' {
		return nil, cursor, errors.ErrExpecting("integer", buf, cursor)
	}
	cursor++

	e := bytes.IndexByte(buf[cursor:], 'e')
	if e == -1 {
		return nil, cursor, errors.ErrSyntax("invalid integer, missing ending char 'e'", cursor)
	}

	if e == 0 {
		return nil, cursor, errors.ErrSyntax("invalid integer", cursor)
	}

	// i ... e

	b := buf[cursor : cursor+e]

	if e == 1 {
		if b[0] < '0' || b[0] > '9' {
			return nil, cursor, errors.ErrSyntax("invalid int", cursor)
		}

		return b, cursor + e + 1, nil
	}

	// e >= 2

	if b[0] == '-' {
		if b[1] == '0' {
			return nil, cursor, errors.ErrSyntax("invalid int '-0' is not allowed", cursor)
		}

		if !validIntBytes(b[1:]) {
			return nil, cursor, errors.ErrSyntax("invalid int", cursor)
		}

		return b, cursor + e + 1, nil
	}

	if b[0] == '0' {
		return nil, cursor, errors.ErrSyntax("invalid int", cursor)
	}

	if !validIntBytes(b) {
		return nil, cursor, errors.ErrSyntax("invalid int", cursor)
	}

	return b, cursor + e + 1, nil
}

func (d *intDecoder) Decode(ctx *Context, cursor int, depth int64, rv reflect.Value) (int, error) {
	buf, c, err := decodeIntegerBytes(ctx.Buf, cursor)
	if err != nil {
		return 0, err
	}

	cursor = c

	return d.processBytes(buf, cursor, rv)
}

func (d *intDecoder) processBytes(bytes []byte, cursor int, rv reflect.Value) (int, error) {
	i64, err := strconv.ParseInt(string(bytes), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to decode int from bencode: %w", err)
	}

	if rv.OverflowInt(i64) {
		return 0, errors.ErrValueOverflow(i64, rv.Type().Kind().String())
	}

	rv.SetInt(i64)

	return cursor, nil
}

func validIntBytes(buf []byte) bool {
	for _, b := range buf[0:] {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}

var typeBigInt = reflect.TypeFor[big.Int]()
var typeBigIntPtr = reflect.TypeFor[*big.Int]()

type bigIntDecoder struct {
	ptrDecoder bigIntPtrDecoder
}

func (b *bigIntDecoder) Decode(ctx *Context, cursor int, depth int64, rv reflect.Value) (int, error) {
	return b.ptrDecoder.Decode(ctx, cursor, depth, rv.Addr())
}

type bigIntPtrDecoder struct {
}

func (b *bigIntPtrDecoder) Decode(ctx *Context, cursor int, depth int64, rv reflect.Value) (int, error) {
	buf, c, err := decodeIntegerBytes(ctx.Buf, cursor)
	if err != nil {
		return 0, err
	}

	cursor = c

	v := rv.Interface().(*big.Int)

	if v == nil {
		v = &big.Int{}
		rv.Set(reflect.ValueOf(v))
	}

	_, ok := v.SetString(string(buf), 10)
	if !ok {
		return 0, errors.ErrSyntax("bencode: invalid int", cursor)
	}

	return c, nil
}
