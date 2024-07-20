package encoder

import (
	"fmt"
	"reflect"
)

type encoder func(ctx *Ctx, b []byte, rv reflect.Value) ([]byte, error)

func compileType(rt reflect.Type) (encoder, error) {
	return compile(rt, seenMap{})
}

func compile(rt reflect.Type, seen seenMap) (encoder, error) {
	if rt.Implements(marshalerType) {
		return compileMarshaler(rt)
	}

	switch rt.Kind() {
	case reflect.Bool:
		return encodeBool, nil
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return encodeInt, nil
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		return encodeUint, nil
	case reflect.String:
		return encodeString, nil
	case reflect.Struct:
		return compileStruct(rt, seen)
	case reflect.Array:
		return compileArray(rt)
	case reflect.Slice:
		return compileSlice(rt, seen)
	case reflect.Map:
		return compileMap(rt, seen)
	case reflect.Interface:
		return compileInterface(rt)
	case reflect.Ptr:
		return compilePtr(rt, seen)
	}

	return nil, fmt.Errorf("failed to build encoder, unsupported type %s (kind %s)", rt.String(), rt.Kind())
}
