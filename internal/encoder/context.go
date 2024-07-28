package encoder

import (
	"sync"
)

var ctxPool = sync.Pool{
	New: func() any {
		return &Context{
			Buf: make([]byte, 0, 8*1024),
		}
	},
}

type Context struct {
	Buf []byte
}

func newCtx() *Context {
	return ctxPool.Get().(*Context)
}

func freeCtx(ctx *Context) {
	if cap(ctx.Buf) >= 100*1024*1024 { // drop buffer that are too long
		return
	}

	ctx.Buf = ctx.Buf[:0]
	ctxPool.Put(ctx)
}
