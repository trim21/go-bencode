package encoder

import (
	"reflect"
)

type UnsupportedTypeAsMapKeyError struct {
	Type reflect.Type
}

func (e *UnsupportedTypeAsMapKeyError) Error() string {
	return "bencode: unsupported type as key of map: " + e.Type.String()
}
