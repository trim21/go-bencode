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

	// Encodes both the field name and value.
	encode    encoder
	fieldName string // field fieldName
	omitEmpty bool
	// support for Anonymous struct
	isZero func(reflect.Value) bool
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
	fields, err := compileStructFieldsEncoders(rt, seen)
	if err != nil {
		return nil, err
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
		// shadow compiler's error
		var err error

		b = append(b, 'd')

		for _, field := range fields {
			v := rv
			for _, index := range field.fieldIndex {
				v = v.Field(index)
			}

			if field.omitEmpty {
				if field.isZero(v) {
					continue
				}
			}

			b, err = field.encode(ctx, b, v)
			if err != nil {
				return b, err
			}
		}

		return append(b, 'e'), nil
	}, nil
}

func compileStructField(rt reflect.Type, fieldName string, seen seenMap) (encoder, error) {
	if rt.Kind() != reflect.Pointer {
		inner, err := compile(rt, seen)
		if err != nil {
			return nil, err
		}

		return func(ctx *Context, b []byte, rv reflect.Value) ([]byte, error) {
			return inner(ctx, AppendStr(b, fieldName), rv)
		}, nil
	}

	if rt.Elem().Kind() == reflect.Pointer {
		return nil, fmt.Errorf("bencode: nested ptr is not supported %s", rt.String())
	}

	inner, err := compile(rt.Elem(), seen)
	if err != nil {
		return nil, err
	}

	elemStruct := rt.Elem().Kind() == reflect.Struct

	return func(ctx *Context, b []byte, rv reflect.Value) ([]byte, error) {
		if rv.IsNil() {
			return b, nil
		}

		b = AppendStr(b, fieldName)

		if elemStruct {
			if ctx.depth++; ctx.depth > startDetectingCyclesAfter {
				ptr := rv.UnsafePointer()
				if _, ok := ctx.ptrSeen[ptr]; ok {
					return b, fmt.Errorf("bencode: encountered a cycle via %s", rv.Type())
				}
				ctx.ptrSeen[ptr] = empty{}
				defer delete(ctx.ptrSeen, ptr)
			}
		}

		b, encodeErr := inner(ctx, b, rv.Elem())

		if elemStruct {
			ctx.depth--
		}

		return b, encodeErr
	}, nil
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

	fieldEncoder, err := compileStructField(rt, cfg.Name(), seen)
	if err != nil {
		return nil, err
	}

	encoders = append(encoders, structEncoder{
		fieldIndex: append(slices.Clone(fieldIndex), index),
		encode:     fieldEncoder,
		fieldName:  cfg.Name(),
		isZero:     compileIsZero(ft.Type),
		omitEmpty:  cfg.IsOmitEmpty,
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

	return func(rv reflect.Value) bool {
		return rv.IsZero()
	}
}
