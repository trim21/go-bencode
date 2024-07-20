package encoder

import (
	"reflect"
)

func reflectSlice(ctx *Ctx, b []byte, rv reflect.Value) ([]byte, error) {
	rt := rv.Type()

	// not slice of interface, fast path
	if rt.Elem().Kind() != reflect.Interface {
		return reflectConcreteSlice(ctx, b, rv)
	}

	if rv.IsNil() {
		return appendEmptyList(b), nil
	}

	enc, err := compileInterface(rt.Elem())
	if err != nil {
		return nil, err
	}

	size := rv.Len()

	for i := 0; i < size; i++ {
		b, err = enc(ctx, b, rv.Index(i))
		if err != nil {
			return b, err
		}
	}

	return append(b, '}'), nil
}

func reflectConcreteSlice(ctx *Ctx, b []byte, rv reflect.Value) ([]byte, error) {
	enc, err := compileWithCache(rv.Type())
	if err != nil {
		return nil, err
	}

	return enc(ctx, b, rv)
}
