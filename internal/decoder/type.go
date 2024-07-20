package decoder

import (
	"reflect"
)

type Decoder interface {
	Decode(ctx *Context, cursor int, depth int64, rv reflect.Value) (int, error)
}

const (
	maxDecodeNestingDepth = 10000
)
