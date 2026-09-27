package decoder

import (
	"bytes"
	"fmt"
	"math"
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

// decodeIntegerBytes decodes the integer at cursor. It returns the digits, the
// absolute value in num, whether the value is negative, and whether it doesn't fit
// in a uint64: big.Int still needs the digits of such a value, while the fixed size
// types report an error. end is the cursor of the value after the integer.
func decodeIntegerBytes(buf []byte, cursor int) (b []byte, num uint64, neg, overflow bool, end int, err error) {
	if buf[cursor] != 'i' {
		return nil, 0, false, false, cursor, errors.ErrExpecting("integer", buf, cursor)
	}
	cursor++

	e := bytes.IndexByte(buf[cursor:], 'e')
	if e == -1 {
		return nil, 0, false, false, cursor, errors.ErrSyntax("invalid integer, missing ending char 'e'", cursor)
	}

	if e == 0 {
		return nil, 0, false, false, cursor, errors.ErrSyntax("invalid integer", cursor)
	}

	// i ... e

	b = buf[cursor : cursor+e]
	end = cursor + e + 1

	digits := b
	if b[0] == '-' {
		digits = b[1:]
		if len(digits) == 0 {
			return nil, 0, false, false, cursor, errors.ErrSyntax("invalid int", cursor)
		}
		if digits[0] == '0' {
			return nil, 0, false, false, cursor, errors.ErrSyntax("invalid int '-0' is not allowed", cursor)
		}
		neg = true
	} else if b[0] == '0' && len(b) > 1 {
		return nil, 0, false, false, cursor, errors.ErrSyntax("invalid int", cursor)
	}

	for _, c := range digits {
		if c < '0' || c > '9' {
			return nil, 0, false, false, cursor, errors.ErrSyntax("invalid int", cursor)
		}

		if overflow {
			continue
		}

		d := uint64(c - '0')
		if num > (math.MaxUint64-d)/10 {
			overflow = true
			continue
		}

		num = num*10 + d
	}

	return b, num, neg, overflow, end, nil
}

// parseInt64 returns the value of the decoded integer, or the error which
// strconv.ParseInt gives for it: a value which doesn't fit in an int64 is an error
// path, and it is reported as it was.
func parseInt64(b []byte, num uint64, neg, overflow bool) (int64, error) {
	if overflow || (neg && num > 1<<63) || (!neg && num > 1<<63-1) {
		return strconv.ParseInt(string(b), 10, 64)
	}

	i64 := int64(num)
	if neg {
		i64 = -i64
	}

	return i64, nil
}

func (d *intDecoder) Decode(ctx *Context, cursor int, depth int64, rv reflect.Value) (int, error) {
	b, num, neg, overflow, c, err := decodeIntegerBytes(ctx.Buf, cursor)
	if err != nil {
		return 0, err
	}

	return d.processBytes(b, num, neg, overflow, c, rv)
}

func (d *intDecoder) processBytes(b []byte, num uint64, neg, overflow bool, cursor int, rv reflect.Value) (int, error) {
	i64, err := parseInt64(b, num, neg, overflow)
	if err != nil {
		return 0, fmt.Errorf("failed to decode int from bencode: %w", err)
	}

	if rv.OverflowInt(i64) {
		return 0, errors.ErrValueOverflow(i64, rv.Type().Kind().String())
	}

	rv.SetInt(i64)

	return cursor, nil
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
	buf, _, _, _, c, err := decodeIntegerBytes(ctx.Buf, cursor)
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
