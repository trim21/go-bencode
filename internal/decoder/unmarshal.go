package decoder

import (
	"fmt"
	"reflect"

	"github.com/trim21/go-bencode/internal/errors"
)

func Unmarshal(data []byte, v any) error {
	return unmarshal(data, v, false)
}

// UnmarshalRelaxed is like Unmarshal but with relaxed parsing rules:
// - Dictionary keys are not required to be sorted
// - Duplicate dictionary keys are allowed (last value wins)
func UnmarshalRelaxed(data []byte, v any) error {
	return unmarshal(data, v, true)
}

func unmarshal(data []byte, v any, relaxed bool) error {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return &errors.InvalidUnmarshalError{}
	}

	rt := rv.Type()

	if err := validateType(rt); err != nil {
		return err
	}
	if rv.IsNil() {
		return &errors.InvalidUnmarshalError{Type: rt}
	}

	dec, err := CompileToGetDecoder(rt)
	if err != nil {
		return err
	}
	ctx := newCtx()
	ctx.Buf = data
	ctx.Relaxed = relaxed
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
	if rt == nil || rt.Kind() != reflect.Pointer {
		return &errors.InvalidUnmarshalError{Type: rt}
	}
	return nil
}
