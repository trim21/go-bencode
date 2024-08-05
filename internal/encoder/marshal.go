package encoder

import (
	"reflect"
)

func MarshalCtx(ctx *Context, v any) error {
	b, err := encode(ctx, ctx.Buf, v)
	if err != nil {
		return err
	}

	ctx.Buf = b

	return nil
}

func encode(ctx *Context, b []byte, v any) ([]byte, error) {
	rv := reflect.ValueOf(v)

	enc, err := compileWithCache(rv.Type())
	if err != nil {
		return nil, err
	}

	return enc(ctx, b, rv)
}
