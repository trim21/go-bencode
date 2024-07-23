package encoder

import (
	"reflect"
)

func compileArray(rt reflect.Type) (encoder, error) {
	size := rt.Len()

	enc, err := compileWithCache(rt.Elem())
	if err != nil {
		return nil, err
	}

	return func(ctx *Context, b []byte, rv reflect.Value) ([]byte, error) {
		b = append(b, 'l')

		var err error // shadow compiler's error
		for i := 0; i < size; i++ {
			b, err = enc(ctx, b, rv.Index(i))
			if err != nil {
				return b, err
			}
		}

		return append(b, 'e'), nil
	}, nil
}
