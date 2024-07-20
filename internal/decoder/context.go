package decoder

import (
	"strconv"
	"sync"
)

type Context struct {
	Buf []byte
}

var (
	runtimeContextPool = sync.Pool{
		New: func() any {
			return &Context{}
		},
	}
)

func TakeRuntimeContext() *Context {
	return runtimeContextPool.Get().(*Context)
}

func ReleaseRuntimeContext(ctx *Context) {
	runtimeContextPool.Put(ctx)
}

var (
	isWhiteSpace = [256]bool{}

	isInteger = [256]bool{}
)

func init() {
	isWhiteSpace[' '] = true
	isWhiteSpace['\n'] = true
	isWhiteSpace['\t'] = true
	isWhiteSpace['\r'] = true

	for i := 0; i < 10; i++ {
		isInteger[[]byte(strconv.Itoa(i))[0]] = true
	}
}
