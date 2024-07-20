package decoder

import (
	"fmt"
	"reflect"
)

var (
	bytesType = reflect.TypeOf([]byte{})
)

type bytesSliceDecoder struct {
	rt         reflect.Type
	structName string
	fieldName  string
}

func newByteSliceDecoder(rt reflect.Type, structName string, fieldName string) Decoder {
	return &bytesSliceDecoder{
		rt:         rt,
		structName: structName,
		fieldName:  fieldName,
	}
}

func (d *bytesSliceDecoder) Decode(ctx *Context, cursor int, depth int64, rv reflect.Value) (int, error) {
	bytes, c, err := readString(ctx.Buf, cursor)
	if err != nil {
		return 0, err
	}

	rv.SetBytes(bytes)
	return c, nil
}

func newByteArrayDecoder(rt reflect.Type, structName string, fieldName string) Decoder {
	return &bytesArrayDecoder{
		rt:         rt,
		size:       rt.Len(),
		structName: structName,
		fieldName:  fieldName,
	}
}

type bytesArrayDecoder struct {
	rt         reflect.Type
	size       int
	structName string
	fieldName  string
}

func (a *bytesArrayDecoder) Decode(ctx *Context, cursor int, depth int64, rv reflect.Value) (int, error) {
	bytes, end, err := readString(ctx.Buf, cursor)
	if err != nil {
		return 0, err
	}

	if len(bytes) != a.size {
		return 0, fmt.Errorf("string length mismatch expected size: expecting %d, actuall %d. index %d", a.size, len(bytes), cursor)
	}

	// SetBytes doesn't work with array bytes
	for i, b := range bytes {
		rv.Index(i).SetUint(uint64(b))
	}

	return end, nil
}
