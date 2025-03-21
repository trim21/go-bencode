package encoder

import (
	"fmt"
	"reflect"
	"sync/atomic"
)

type encoder func(ctx *Context, b []byte, rv reflect.Value) ([]byte, error)

var cachedEncoderMap atomic.Pointer[map[reflect.Type]encoder]

func init() {
	cachedEncoderMap.Store(&map[reflect.Type]encoder{})
}

func compileWithCache(rt reflect.Type) (encoder, error) {
	opcodeMap := *cachedEncoderMap.Load()
	if codeSet, exists := opcodeMap[rt]; exists {
		return codeSet, nil
	}
	codeSet, err := compile(rt, seenMap{})
	if err != nil {
		return nil, err
	}
	storeEncoder(rt, codeSet, opcodeMap)
	return codeSet, nil
}

func storeEncoder(rt reflect.Type, set encoder, m map[reflect.Type]encoder) {
	newEncoderMap := make(map[reflect.Type]encoder, len(m)+1)
	newEncoderMap[rt] = set

	for k, v := range m {
		newEncoderMap[k] = v
	}

	cachedEncoderMap.Store(&newEncoderMap)
}

func compile(rt reflect.Type, seen seenMap) (encoder, error) {
	switch {
	case rt.Implements(marshalerType):
		return compileMarshaler(rt)
	case rt == bytesType:
		return encodeBytesSlice, nil
	case rt.Kind() == reflect.Array && rt.Elem().Kind() == reflect.Uint8:
		return compileBytesArray(rt)
	case rt == typeBigInt:
		return encodeBigInt, nil
	case rt == typeBigIntPtr:
		return encodeBigIntPtr, nil
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
	default:
		return nil, fmt.Errorf("failed to build encoder, unsupported type %s (kind %s)", rt.String(), rt.Kind())
	}
}
