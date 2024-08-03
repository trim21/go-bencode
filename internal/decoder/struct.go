package decoder

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/trim21/go-bencode/internal/errors"
	"github.com/trim21/go-bencode/internal/runtime"
)

func compileStruct(rt reflect.Type, structName, fieldName string, structTypeToDecoder map[reflect.Type]Decoder) (Decoder, error) {
	if dec, exists := structTypeToDecoder[rt]; exists {
		return dec, nil
	}
	structDec := newStructDecoder(structName, fieldName, map[string]*structFieldDecoder{})
	structDec.structName = rt.Name()
	structTypeToDecoder[rt] = structDec
	structName = rt.Name()

	fieldNum := rt.NumField()

	var allFields = make([]*structFieldDecoder, 0, fieldNum)

	for i := 0; i < fieldNum; i++ {
		field := rt.Field(i)
		tag := runtime.StructTagFromField(field)
		if tag.Key == "-" || runtime.IsIgnoredStructField(field) {
			continue
		}

		var key string
		if tag.Key != "" {
			key = tag.Key
		} else {
			key = field.Name
		}

		if field.Anonymous {
			if rt.Kind() != reflect.Struct {
				return nil, fmt.Errorf("bencode: only support struct as Anonymous field, found: %s", rt.String())
			}

			if field.Tag.Get("bencode") == "" {
				enc, err := compileStruct(field.Type, structName, key, structTypeToDecoder)
				if err != nil {
					return nil, err
				}

				se := enc.(*structDecoder)
				for _, dec := range se.fieldMap {
					allFields = append(allFields, &structFieldDecoder{
						dec:        dec.dec,
						fieldIndex: append([]int{i}, dec.fieldIndex...),
						key:        dec.key,
					})
				}
				continue
			}
		}

		dec, err := compile(field.Type, structName, key, structTypeToDecoder)
		if err != nil {
			return nil, err
		}

		fieldSet := &structFieldDecoder{
			dec:        dec,
			fieldIndex: []int{i},
			key:        key,
		}

		allFields = append(allFields, fieldSet)
	}

	seen := map[string]bool{}
	for _, dec := range allFields {
		if seen[dec.key] {
			return nil, fmt.Errorf("found duplicate keys in struct %s: %s", rt.String(), dec.key)
		}

		seen[dec.key] = true
		structDec.fieldMap[dec.key] = dec
	}

	delete(structTypeToDecoder, rt)

	return structDec, nil
}

type structFieldDecoder struct {
	key string

	dec Decoder

	fieldIndex []int // for anonymous struct field
}

type structDecoder struct {
	fieldMap   map[string]*structFieldDecoder
	structName string
	fieldName  string
}

func newStructDecoder(structName, fieldName string, fieldMap map[string]*structFieldDecoder) *structDecoder {
	return &structDecoder{
		fieldMap:   fieldMap,
		structName: structName,
		fieldName:  fieldName,
	}
}

func decodeKey(d *structDecoder, buf []byte, cursor int) ([]byte, int, *structFieldDecoder, error) {
	key, c, err := readString(buf, cursor)
	if err != nil {
		return nil, 0, nil, err
	}

	// go compiler will not alloc key in this case
	field := d.fieldMap[string(key)]

	return key, c, field, nil
}

func (d *structDecoder) Decode(ctx *Context, cursor int, depth int64, rv reflect.Value) (int, error) {
	buf := ctx.Buf
	bufSize := len(buf)
	if cursor+2 > bufSize {
		return 0, errors.ErrSyntax("buffer overflow when parsing directory", cursor)
	}

	depth++
	if depth > maxDecodeNestingDepth {
		return 0, errors.ErrExceededMaxDepth(ctx.Buf[cursor], cursor)
	}

	if buf[cursor] != 'd' {
		return 0, errors.ErrInvalidBeginningOfValue(buf[cursor], cursor)
	}

	cursor++

	var lastKey []byte

	for {
		if cursor >= bufSize {
			return 0, errors.DataTooShort()
		}

		if buf[cursor] == 'e' {
			cursor++
			return cursor, nil
		}

		currentKey, c, field, err := decodeKey(d, buf, cursor)
		if err != nil {
			return 0, err
		}

		if lastKey != nil {
			switch bytes.Compare(lastKey, currentKey) {
			case 0:
				return cursor, fmt.Errorf("dictionary conrains duplicated keys %s. index %d", currentKey, cursor)
			case 1:
				return cursor, fmt.Errorf("dictionary conrains unordered keys %s, %s. index %d", lastKey, currentKey, cursor)
			}
		}
		lastKey = currentKey

		cursor = c

		if cursor >= bufSize {
			return 0, errors.ErrExpecting("object value after colon", buf, cursor)
		}

		if field == nil {
			cursor, err = skipValue(buf, cursor, depth)
			if err != nil {
				return 0, err
			}
			continue
		}

		v := rv
		for _, index := range field.fieldIndex {
			v = v.Field(index)
		}

		cursor, err = field.dec.Decode(ctx, cursor, depth, v)
		if err != nil {
			return 0, fmt.Errorf("bencode: failed to decode Go struct field %s.%s: %w", d.structName, field.key, err)
		}
	}
}
