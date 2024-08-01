package encoder

import (
	"reflect"
)

func reflectMap(ctx *Context, b []byte, rv reflect.Value) ([]byte, error) {
	if rv.IsNil() {
		return appendEmptyMap(b), nil
	}

	rt := rv.Type()

	enc, err := compileWithCache(rt)
	if err != nil {
		return nil, err
	}

	return enc(ctx, b, rv)
}
