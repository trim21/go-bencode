package decoder

import (
	"reflect"

	"github.com/trim21/go-bencode/internal/errors"
)

type interfaceDecoder struct {
	rt               reflect.Type
	structName       string
	fieldName        string
	sliceDecoder     *sliceDecoder
	mapDecoder       *mapDecoder
	stringDecoder    *stringDecoder
	intDecode        *intDecoder
	mapAnyKeyDecoder *mapKeyDecoder
}

func newEmptyInterfaceDecoder(structName, fieldName string) *interfaceDecoder {
	ifaceDecoder := &interfaceDecoder{
		rt:            emptyInterfaceType,
		structName:    structName,
		fieldName:     fieldName,
		intDecode:     newIntDecoder(interfaceIntType, structName, fieldName),
		stringDecoder: newStringDecoder(structName, fieldName),
	}

	ifaceDecoder.mapAnyKeyDecoder = newInterfaceMapKeyDecoder(ifaceDecoder.stringDecoder)

	ifaceDecoder.sliceDecoder = newSliceDecoder(
		ifaceDecoder,
		emptyInterfaceType,
		structName, fieldName,
	)

	ifaceDecoder.mapDecoder = newMapDecoder(
		interfaceClassMapType,
		stringType,
		ifaceDecoder.stringDecoder,
		interfaceClassMapType.Elem(),
		ifaceDecoder,
		structName,
		fieldName,
	)

	return ifaceDecoder
}

func newInterfaceDecoder(rt reflect.Type, structName, fieldName string) *interfaceDecoder {
	emptyIfaceDecoder := newEmptyInterfaceDecoder(structName, fieldName)
	return &interfaceDecoder{
		rt:         rt,
		structName: structName,
		fieldName:  fieldName,
		sliceDecoder: newSliceDecoder(
			emptyIfaceDecoder,
			emptyInterfaceType,
			structName, fieldName,
		),
		stringDecoder:    newStringDecoder(structName, fieldName),
		intDecode:        emptyIfaceDecoder.intDecode,
		mapDecoder:       emptyIfaceDecoder.mapDecoder,
		mapAnyKeyDecoder: emptyIfaceDecoder.mapAnyKeyDecoder,
	}
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
	if rv.NumMethod() > 0 && rv.CanInterface() {
		if u, ok := rv.Interface().(Unmarshaler); ok {
			return decodeUnmarshaler(buf, cursor, depth, u)
		}
		return 0, d.errUnmarshalType(rv.Type(), cursor)
	}

	if rv.Type().NumMethod() == 0 {
		// concrete type is empty interface
		return d.decodeEmptyInterface(ctx, cursor, depth, rv)
	}
	if rv.Type().Kind() == reflect.Ptr && rv.Type().Elem() == d.rt || rv.Type().Kind() != reflect.Ptr {
		return d.decodeEmptyInterface(ctx, cursor, depth, rv)
	}
	decoder, err := CompileToGetDecoder(rv.Type())
	if err != nil {
		return 0, err
	}
	return decoder.Decode(ctx, cursor, depth, rv)
}

func (d *interfaceDecoder) decodeEmptyInterface(ctx *Context, cursor int, depth int64, rv reflect.Value) (int, error) {
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
		return 0, errors.ErrExpected("array key", cursor)
	}
}

func newInterfaceMapKeyDecoder(stringDecoder *stringDecoder) *mapKeyDecoder {
	return &mapKeyDecoder{strDecoder: stringDecoder}
}
