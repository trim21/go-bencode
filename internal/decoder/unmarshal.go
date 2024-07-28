package decoder

import (
	"fmt"
	"reflect"

	"github.com/trim21/go-bencode/internal/errors"
)

func Unmarshal(data []byte, v any) error {
	rv := reflect.ValueOf(v)

	rt := rv.Type()

	if err := validateType(rt); err != nil {
		return err
	}

	dec, err := CompileToGetDecoder(rt)
	if err != nil {
		return err
	}
	ctx := newCtx()
	ctx.Buf = data
	cursor, err := dec.Decode(ctx, 0, 0, rv.Elem())
	if err != nil {
		freeCtx(ctx)
		return err
	}
	freeCtx(ctx)
	return validateEndBuf(data, cursor)
}

func validateEndBuf(src []byte, cursor int) error {
	if len(src) == cursor {
		return nil
	}

	return errors.ErrSyntax(
		fmt.Sprintf("invalid character '%c' after top-level value", src[cursor]),
		cursor+1,
	)
}

func validateType(rt reflect.Type) error {
	if rt == nil || rt.Kind() != reflect.Ptr {
		return &errors.InvalidUnmarshalError{Type: rt}
	}
	return nil
}
