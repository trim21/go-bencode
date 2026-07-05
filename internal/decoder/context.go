package decoder

import (
	"sync"
)

type Context struct {
	Buf     []byte
	Relaxed bool
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
	ctx.Relaxed = false
	ctxPool.Put(ctx)
}
