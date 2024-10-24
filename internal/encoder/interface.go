package encoder

import (
	"reflect"
)

// will need to get type message at marshal time, slow path.
// should avoid interface for performance thinking.
func compileInterface(rt reflect.Type) (encoder, error) {
	return reflectInterfaceValue, nil
}

func reflectInterfaceValue(ctx *Context, b []byte, rv reflect.Value) ([]byte, error) {
LOOP:
	for {
		switch rv.Kind() {
		case reflect.Ptr, reflect.Interface:
			if rv.IsNil() || rv.IsZero() {
				return b, nil
			}
			rv = rv.Elem()
		default:
			break LOOP
		}
	}

	// simple type
	switch rv.Kind() {
	case reflect.Bool:
		return encodeBool(ctx, b, rv)
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		return encodeUint(ctx, b, rv)
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return encodeInt(ctx, b, rv)
	case reflect.String:
		return encodeString(ctx, b, rv)
	}

	enc, err := compileWithCache(rv.Type())
	if err != nil {
		return nil, err
	}
	return enc(ctx, b, rv)
}
