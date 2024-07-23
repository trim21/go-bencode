package encoder

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/trim21/go-bencode/internal/runtime"
)

type structEncoder struct {
	index int
	// a direct value handler, like `encodeInt`
	// struct encoder should de-ref pointers and pass real address to encoder.
	// address of map, slice, array may still be 0, bug theirs encoder will handle that at null.
	encode    encoder
	fieldName string // field fieldName
	omitEmpty bool
	isZero    func(reflect.Value) bool
	ptr       bool
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
	} else {
		seen[rt] = &structRecEncoder{}
	}

	enc, err := compileStructFields(rt, seen)
	if err != nil {
		return nil, err
	}

	recursiveEnc, recursiveStruct := seen[rt]
	if recursiveStruct {
		if recursiveEnc.enc == nil {
			recursiveEnc.enc = enc
			return recursiveEnc.Encode, nil
		}
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
			return nil, fmt.Errorf("duplicate field name %s", field.fieldName)
		}
		fieldNames[field.fieldName] = true
	}

	return func(ctx *Context, b []byte, rv reflect.Value) ([]byte, error) {
		// shadow compiler's error
		var err error

		b = append(b, 'd')

		for _, field := range fields {
			v := rv.Field(field.index)

			if field.omitEmpty {
				if field.isZero(v) {
					continue
				}
			}

			if field.ptr {
				if v.IsNil() {
					continue
				}

				b = appendString(b, field.fieldName)
				b, err = field.encode(ctx, b, v.Elem())
				if err != nil {
					return b, err
				}
				continue
			}

			b = appendString(b, field.fieldName)
			b, err = field.encode(ctx, b, v)
			if err != nil {
				return b, err
			}
		}

		return append(b, 'e'), nil
	}, nil
}

func compileStructFieldsEncoders(rt reflect.Type, seen seenMap) ([]structEncoder, error) {
	var encoders []structEncoder

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		cfg := runtime.StructTagFromField(field)
		if cfg.Key == "-" || !cfg.Field.IsExported() {
			continue
		}

		var fieldEncoder encoder
		var err error

		var isPtrField = field.Type.Kind() == reflect.Ptr

		if field.Type.Kind() == reflect.Ptr {
			if field.Type.Elem().Kind() == reflect.Ptr {
				return nil, fmt.Errorf("encoding nested ptr is not supported %s", field.Type.String())
			}
		}

		if field.Anonymous {
			if field.Type.Kind() == reflect.Struct || (field.Type.Kind() == reflect.Ptr && field.Type.Kind() == reflect.Struct) {
				return nil, fmt.Errorf("supported for Anonymous struct field has been removed: %s", field.Type.String())
			}
		}

		if field.Type.Kind() == reflect.Ptr {
			fieldEncoder, err = compile(field.Type.Elem(), seen)
		} else {
			fieldEncoder, err = compile(field.Type, seen)
		}

		if err != nil {
			return nil, err
		}

		encoders = append(encoders, structEncoder{
			index:     i,
			encode:    fieldEncoder,
			fieldName: cfg.Name(),
			isZero:    compileIsZero(field.Type),
			omitEmpty: cfg.IsOmitEmpty,
			ptr:       isPtrField,
		})
	}

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
