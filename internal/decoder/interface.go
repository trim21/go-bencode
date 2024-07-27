package decoder

import (
	"reflect"

	"github.com/trim21/go-bencode/internal/errors"
)

type interfaceDecoder struct {
	rt            reflect.Type
	structName    string
	fieldName     string
	sliceDecoder  *sliceDecoder
	mapDecoder    *mapDecoder
	stringDecoder *stringDecoder
	intDecode     *intDecoder
}

func newEmptyInterfaceDecoder(structName, fieldName string) *interfaceDecoder {
	ifceDecoder := &interfaceDecoder{
		rt:            emptyInterfaceType,
		structName:    structName,
		fieldName:     fieldName,
		intDecode:     newIntDecoder(interfaceIntType, structName, fieldName),
		stringDecoder: newStringDecoder(structName, fieldName),
	}

	ifceDecoder.sliceDecoder = newSliceDecoder(
		ifceDecoder,
		emptyInterfaceType,
		structName,
		fieldName,
	)

	ifceDecoder.mapDecoder = newMapDecoder(
		interfaceClassMapType,
		stringType,
		ifceDecoder.stringDecoder,
		emptyInterfaceType,
		ifceDecoder,
		structName,
		fieldName,
	)

	return ifceDecoder
}

func newInterfaceDecoder(rt reflect.Type, structName, fieldName string) *interfaceDecoder {
	return newEmptyInterfaceDecoder(structName, fieldName)
}

var (
	stringType            = reflect.TypeFor[string]()
	emptyInterfaceType    = reflect.TypeFor[any]()
	interfaceClassMapType = reflect.TypeFor[map[string]any]()
	interfaceIntType      = reflect.TypeFor[int64]()
)

func decodeUnmarshaler(buf []byte, cursor int, depth int64, unmarshaler Unmarshaler) (int, error) {
	start := cursor
	end, err := skipValue(buf, cursor, depth)
	if err != nil {
		return 0, err
	}
	src := buf[start:end]
	if err := unmarshaler.UnmarshalBencode(src); err != nil {
		return 0, err
	}
	return end, nil
}

func (d *interfaceDecoder) errUnmarshalType(rt reflect.Type, offset int) *errors.UnmarshalTypeError {
	return &errors.UnmarshalTypeError{
		Value:  rt.String(),
		Type:   rt,
		Offset: offset,
		Struct: d.structName,
		Field:  d.fieldName,
	}
}

func (d *interfaceDecoder) Decode(ctx *Context, cursor int, depth int64, rv reflect.Value) (int, error) {
	buf := ctx.Buf
	if cursor >= len(buf) {
		return 0, errors.ErrSyntax("input too short when decoding any", cursor)
	}

	switch buf[cursor] {
	case 'd':
		var v map[string]any
		value := reflect.ValueOf(&v).Elem()
		end, err := d.mapDecoder.Decode(ctx, cursor, depth, value)
		if err != nil {
			return 0, err
		}
		rv.Set(value)
		return end, nil
	case 'l':
		var v []any
		value := reflect.ValueOf(&v).Elem()
		end, err := d.sliceDecoder.Decode(ctx, cursor, depth, value)
		if err != nil {
			return 0, err
		}
		rv.Set(value)
		return end, nil
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		b, end, err := readString(buf, cursor)
		if err != nil {
			return 0, err
		}
		rv.Set(reflect.ValueOf(string(b)))
		return end, nil
	case 'i':
		var v int64
		value := reflect.ValueOf(&v).Elem()
		end, err := d.intDecode.Decode(ctx, cursor, depth, value)
		if err != nil {
			return 0, err
		}
		rv.Set(value)
		return end, nil
	}

	return cursor, errors.ErrInvalidBeginningOfValue(buf[cursor], cursor)
}

type mapKeyDecoder struct {
	strDecoder *stringDecoder
}

func (d *mapKeyDecoder) Decode(ctx *Context, cursor int, depth int64, rv reflect.Value) (int, error) {
	buf := ctx.Buf

	switch buf[cursor] {
	case 's':
		var v string
		ptr := reflect.ValueOf(&v).Elem()
		cursor, err := d.strDecoder.Decode(ctx, cursor, depth, ptr)
		if err != nil {
			return 0, err
		}
		rv.Set(ptr)
		return cursor, nil
	// string key
	default:
		return 0, errors.ErrExpecting("array key", buf, cursor)
	}
}
