package decoder

import (
	"fmt"
	"reflect"

	"github.com/trim21/go-bencode/internal/errors"
	"github.com/trim21/go-bencode/internal/runtime"
)

func compileStruct(rt reflect.Type, structName, fieldName string, structTypeToDecoder map[reflect.Type]Decoder) (Decoder, error) {
	if dec, exists := structTypeToDecoder[rt]; exists {
		return dec, nil
	}
	structDec := newStructDecoder(structName, fieldName, map[string]*structFieldSet{})
	structTypeToDecoder[rt] = structDec
	structName = rt.Name()

	var allFields []*structFieldSet

	fieldNum := rt.NumField()
	for i := 0; i < fieldNum; i++ {
		field := rt.Field(i)
		if runtime.IsIgnoredStructField(field) {
			continue
		}

		if field.Anonymous {
			if (field.Type.Kind() == reflect.Struct) || (field.Type.Kind() == reflect.Ptr && (field.Type.Elem().Kind() == reflect.Struct)) {
				return nil, fmt.Errorf("anonymous struct field is not supported: %s", rt.String())
			}
		}

		tag := runtime.StructTagFromField(field)
		dec, err := compile(field.Type, structName, field.Name, structTypeToDecoder)
		if err != nil {
			return nil, err
		}

		var key string
		if tag.Key != "" {
			key = tag.Key
		} else {
			key = field.Name
		}

		fieldSet := &structFieldSet{
			dec:      dec,
			fieldIdx: i,
			key:      key,
		}

		allFields = append(allFields, fieldSet)
	}

	seen := map[string]bool{}
	for _, set := range allFields {
		if seen[set.key] {
			return nil, fmt.Errorf("found duplicate keys for struct %s: %s", rt.String(), set.key)
		}

		seen[set.key] = true
		structDec.fieldMap[set.key] = set
	}

	delete(structTypeToDecoder, rt)

	return structDec, nil
}

type structFieldSet struct {
	dec      Decoder
	fieldIdx int
	key      string
	err      error
}

type structDecoder struct {
	fieldMap   map[string]*structFieldSet
	structName string
	fieldName  string
}

func newStructDecoder(structName, fieldName string, fieldMap map[string]*structFieldSet) *structDecoder {
	return &structDecoder{
		fieldMap:   fieldMap,
		structName: structName,
		fieldName:  fieldName,
	}
}

func decodeKey(d *structDecoder, buf []byte, cursor int) (int, *structFieldSet, error) {
	key, c, err := readString(buf, cursor)
	if err != nil {
		return 0, nil, err
	}

	// go compiler will not escape key
	field, exists := d.fieldMap[string(key)]
	if !exists {
		return c, nil, nil
	}

	return c, field, nil
}

func (d *structDecoder) Decode(ctx *Context, cursor int, depth int64, rv reflect.Value) (int, error) {
	depth++
	if depth > maxDecodeNestingDepth {
		return 0, errors.ErrExceededMaxDepth(ctx.Buf[cursor], cursor)
	}

	buf := ctx.Buf

	bufSize := len(buf)

	if cursor+2 > bufSize {
		return 0, errors.ErrSyntax("buffer overflow when parsing directory", cursor)
	}

	if buf[cursor] == 'd' {
		cursor++
	} else {
		return 0, errors.ErrInvalidBeginningOfValue(buf[cursor], cursor)
	}

	for {
		if cursor >= bufSize {
			return 0, fmt.Errorf("buffer overflow when decoding dictionary: %d", cursor)
		}

		if buf[cursor] == 'e' {
			cursor++
			return cursor, nil
		}

		c, field, err := decodeKey(d, buf, cursor)
		if err != nil {
			return 0, err
		}

		cursor = c

		if cursor >= bufSize {
			return 0, errors.ErrExpected("object value after colon", cursor)
		}

		if field == nil {
			cursor, err = skipValue(buf, cursor, depth)
			if err != nil {
				return 0, err
			}
			continue
		}

		if field.err != nil {
			return 0, field.err
		}

		cursor, err = field.dec.Decode(ctx, cursor, depth, rv.Field(field.fieldIdx))
		if err != nil {
			return 0, err
		}
	}
}
