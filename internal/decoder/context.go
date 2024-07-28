package decoder

import (
	"sync"
)

type Context struct {
	Buf []byte
}

var ctxPool = sync.Pool{
	New: func() any {
		return &Context{}
	},
}

func newCtx() *Context {
	return ctxPool.Get().(*Context)
}

func freeCtx(ctx *Context) {
	ctx.Buf = nil
	ctxPool.Put(ctx)
}
