package decoder

import (
	"fmt"
	"reflect"
	"sync/atomic"
)

var (
	cachedDecoderMap atomic.Pointer[map[reflect.Type]Decoder]
)

func init() {
	var m = map[reflect.Type]Decoder{}
	cachedDecoderMap.Store(&m)
}

func CompileToGetDecoder(rt reflect.Type) (Decoder, error) {
	decoderMap := *cachedDecoderMap.Load()
	if dec, exists := decoderMap[rt]; exists {
		return dec, nil
	}

	dec, err := compileHead(rt, map[reflect.Type]Decoder{})
	if err != nil {
		return nil, err
	}

	storeDecoder(rt, dec, decoderMap)

	return dec, nil
}

func storeDecoder(rt reflect.Type, dec Decoder, m map[reflect.Type]Decoder) {
	newDecoderMap := make(map[reflect.Type]Decoder, len(m)+1)
	for k, v := range m {
		newDecoderMap[k] = v
	}

	newDecoderMap[rt] = dec

	cachedDecoderMap.Store(&newDecoderMap)
}

func compileHead(rt reflect.Type, structTypeToDecoder map[reflect.Type]Decoder) (Decoder, error) {
	if reflect.PointerTo(rt).Implements(unmarshalerType) {
		return newUnmarshalerDecoder(reflect.PointerTo(rt), "", ""), nil
	}
	return compile(rt.Elem(), "", "", structTypeToDecoder)
}

func compile(rt reflect.Type, structName, fieldName string, structTypeToDecoder map[reflect.Type]Decoder) (Decoder, error) {
	switch {
	case reflect.PointerTo(rt).Implements(unmarshalerType):
		return newUnmarshalerDecoder(reflect.PointerTo(rt), structName, fieldName), nil
	case rt == bytesType:
		return newByteSliceDecoder(rt, structName, fieldName), nil
	case rt.Kind() == reflect.Array && rt.Elem().Kind() == reflect.Uint8:
		return newByteArrayDecoder(rt, structName, fieldName), nil
	}

	switch rt.Kind() {
	case reflect.Ptr:
		return compilePtr(rt, structName, fieldName, structTypeToDecoder)
	case reflect.Struct:
		return compileStruct(rt, structName, fieldName, structTypeToDecoder)
	case reflect.Slice:
		return compileSlice(rt, structName, fieldName, structTypeToDecoder)
	case reflect.Array:
		return compileArray(rt, structName, fieldName, structTypeToDecoder)
	case reflect.Map:
		return compileMap(rt, structName, fieldName, structTypeToDecoder)
	case reflect.Interface:
		return compileInterface(rt, structName, fieldName)
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return newIntDecoder(rt, structName, fieldName), nil
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		return newUintDecoder(rt, structName, fieldName), nil
	case reflect.String:
		return compileString(rt, structName, fieldName)
	case reflect.Bool:
		return compileBool(structName, fieldName)
	}

	return newInvalidDecoder(rt, structName, fieldName), nil
}

func compileMapKey(rt reflect.Type, structName, fieldName string, structTypeToDecoder map[reflect.Type]Decoder) (Decoder, error) {
	switch {
	case reflect.PointerTo(rt).Implements(unmarshalerType):
		return newUnmarshalerDecoder(reflect.PointerTo(rt), structName, fieldName), nil
	case rt.Kind() == reflect.String:
		return newStringDecoder(structName, fieldName), nil
	case rt.Kind() == reflect.Array && rt.Elem().Kind() == reflect.Uint8:
		return newByteArrayDecoder(rt, structName, fieldName), nil
	default:
		return nil, fmt.Errorf("bencode only support [...]byte or string as map key")
	}
}

func compilePtr(rt reflect.Type, structName, fieldName string, structTypeToDecoder map[reflect.Type]Decoder) (Decoder, error) {
	dec, err := compile(rt.Elem(), structName, fieldName, structTypeToDecoder)
	if err != nil {
		return nil, err
	}
	return newPtrDecoder(dec, rt.Elem(), structName, fieldName)
}

func compileString(rt reflect.Type, structName, fieldName string) (Decoder, error) {
	return newStringDecoder(structName, fieldName), nil
}

func compileBool(structName, fieldName string) (Decoder, error) {
	return newBoolDecoder(structName, fieldName), nil
}

func compileSlice(rt reflect.Type, structName, fieldName string, structTypeToDecoder map[reflect.Type]Decoder) (Decoder, error) {
	elem := rt.Elem()
	decoder, err := compile(elem, structName, fieldName, structTypeToDecoder)
	if err != nil {
		return nil, err
	}
	return newSliceDecoder(decoder, elem, structName, fieldName), nil
}

func compileArray(rt reflect.Type, structName, fieldName string, structTypeToDecoder map[reflect.Type]Decoder) (Decoder, error) {
	elem := rt.Elem()
	decoder, err := compile(elem, structName, fieldName, structTypeToDecoder)
	if err != nil {
		return nil, err
	}
	return newArrayDecoder(decoder, elem, rt.Len(), structName, fieldName), nil
}

func compileInterface(rt reflect.Type, structName, fieldName string) (Decoder, error) {
	if rt.NumMethod() != 0 {
		return nil, fmt.Errorf("bencode: can only decode to empty interface, %s is not supported", rt.String())
	}

	return newInterfaceDecoder(rt, structName, fieldName), nil
}
