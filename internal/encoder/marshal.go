package encoder

import (
	"reflect"
)

func MarshalCtx(ctx *Context, v any) error {
	rv := reflect.ValueOf(v)

	enc, err := compileWithCache(rv.Type())
	if err != nil {
		return err
	}

	ctx.Buf, err = enc(ctx, ctx.Buf, rv)

	return err
}
