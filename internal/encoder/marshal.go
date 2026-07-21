package encoder

import (
	"errors"
	"reflect"
)

var ErrNilValue = errors.New("bencode: nil values cannot be encoded")

func MarshalCtx(ctx *Context, v any) error {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return ErrNilValue
	}

	enc, err := compileWithCache(rv.Type())
	if err != nil {
		return err
	}

	ctx.Buf, err = enc(ctx, ctx.Buf, rv)

	return err
}
