package encoder

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/trim21/go-bencode/internal/runtime"
)

type structEncoder struct {
	fieldIndex []int

	// a direct value handler, like `encodeInt`
	// struct encoder should de-ref pointers and pass real address to encoder.
	encode    encoder
	fieldName string // field fieldName
	omitEmpty bool
	// support for Anonymous struct
	isZero func(reflect.Value) bool
	ptr    bool
}

type seenMap = map[reflect.Type]*structRecEncoder

type structRecEncoder struct {
	enc encoder
}

func (s *structRecEncoder) Encode(ctx *Context, b []byte, rv reflect.Value) ([]byte, error) {
	return s.enc(ctx, b, rv)
}

func compileStruct(rt reflect.Type, seen seenMap) (encoder, error) {
	recursiveEnc, hasSeen := seen[rt]

	if hasSeen {
		return recursiveEnc.Encode, nil
	}

	typeEncoder := &structRecEncoder{}

	seen[rt] = typeEncoder

	enc, err := compileStructFields(rt, seen)
	if err != nil {
		return nil, err
	}

	if typeEncoder.enc == nil {
		typeEncoder.enc = enc
		return typeEncoder.Encode, nil
	}

	return enc, nil
}

// struct don't have `omitempty` tag, fast path
func compileStructFields(rt reflect.Type, seen seenMap) (encoder, error) {
	fields, ce := compileStructFieldsEncoders(rt, seen)
	if ce != nil {
		return nil, ce
	}

	slices.SortFunc(fields, func(a, b structEncoder) int {
		return strings.Compare(a.fieldName, b.fieldName)
	})

	var fieldNames = make(map[string]bool, len(fields))

	for _, field := range fields {
		if fieldNames[field.fieldName] {
			return nil, fmt.Errorf("bencode: duplicate field name %s", field.fieldName)
		}
		fieldNames[field.fieldName] = true
	}

	if len(fields) == 0 {
		return func(ctx *Context, b []byte, rv reflect.Value) ([]byte, error) {
			return appendEmptyMap(b), nil
		}, nil
	}

	return func(ctx *Context, b []byte, rv reflect.Value) ([]byte, error) {
		b = append(b, 'd')

		var err error
		for _, field := range fields {
			b, err = encodeStructField(ctx, b, rv, field)
			if err != nil {
				return b, err
			}
		}

		return append(b, 'e'), nil
	}, nil
}

func encodeStructField(ctx *Context, b []byte, rv reflect.Value, field structEncoder) ([]byte, error) {
	var err error
	v := rv
	for _, index := range field.fieldIndex {
		v = v.Field(index)
	}

	if field.omitEmpty {
		if field.isZero(v) {
			return b, nil
		}
	}

	if field.ptr {
		if v.IsNil() {
			return b, nil
		}

		if ctx.ptrLevel++; ctx.ptrLevel > startDetectingCyclesAfter {
			ptr := v.UnsafePointer()
			if _, ok := ctx.ptrSeen[ptr]; ok {
				return b, fmt.Errorf("bencode: encountered a cycle via %s", rv.Type())
			}
			ctx.ptrSeen[ptr] = empty{}
			defer delete(ctx.ptrSeen, ptr)
		}

		v = v.Elem()
	}

	b = AppendStr(b, field.fieldName)
	b, err = field.encode(ctx, b, v)

	if field.ptr {
		ctx.ptrLevel--
	}

	return b, err
}

func compileStructFieldsEncoders(rt reflect.Type, seen seenMap) ([]structEncoder, error) {
	var encoders []structEncoder

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		enc, err := compileStructFieldsEncoder(field, []int{}, i, seen)
		if err != nil {
			return nil, err
		}

		encoders = append(encoders, enc...)
	}

	return encoders, nil
}

func compileStructFieldsEncoder(ft reflect.StructField, fieldIndex []int, index int, seen seenMap) ([]structEncoder, error) {
	cfg := runtime.StructTagFromField(ft)
	if cfg.Key == "-" || !cfg.Field.IsExported() {
		return nil, nil
	}

	rt := ft.Type

	var encoders []structEncoder

	// Do not take struct { S `bencode:n` } as anonymous field
	if ft.Anonymous {
		if rt.Kind() != reflect.Struct {
			return nil, fmt.Errorf("bencode: only support struct as Anonymous field, found: %s", rt.String())
		}

		if ft.Tag.Get("bencode") == "" {
			for ni := 0; ni < rt.NumField(); ni++ {
				nField := rt.Field(ni)
				enc, err := compileStructFieldsEncoder(nField, append(slices.Clone(fieldIndex), index), ni, seen)
				if err != nil {
					return nil, err
				}
				encoders = append(encoders, enc...)
			}

			return encoders, nil
		}
	}

	var fieldEncoder encoder
	var err error

	var isPtrField = rt.Kind() == reflect.Ptr
	if isPtrField {
		if rt.Elem().Kind() == reflect.Ptr {
			return nil, fmt.Errorf("bencode: nested ptr is not supported %s", rt.String())
		}
	}

	if isPtrField {
		fieldEncoder, err = compile(rt.Elem(), seen)
	} else {
		fieldEncoder, err = compile(rt, seen)
	}

	if err != nil {
		return nil, err
	}

	encoders = append(encoders, structEncoder{
		fieldIndex: append(slices.Clone(fieldIndex), index),
		encode:     fieldEncoder,
		fieldName:  cfg.Name(),
		isZero:     compileIsZero(ft.Type),
		omitEmpty:  cfg.IsOmitEmpty,
		ptr:        isPtrField,
	})

	return encoders, nil
}

type IsZeroValue interface {
	IsZeroBencodeValue() bool
}

var isZeroValueType = reflect.TypeFor[IsZeroValue]()

func compileIsZero(rt reflect.Type) func(rv reflect.Value) bool {
	if rt.Implements(isZeroValueType) {
		return func(rv reflect.Value) bool {
			return rv.Interface().(IsZeroValue).IsZeroBencodeValue()
		}
	}

	switch rt.Kind() {
	case reflect.Slice, reflect.Map:
		return func(rv reflect.Value) bool {
			return rv.Len() == 0
		}
	}

	return func(rv reflect.Value) bool {
		return rv.IsZero()
	}
}
