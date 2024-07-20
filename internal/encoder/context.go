package encoder

import (
	"sync"
)

var ctxPool = sync.Pool{
	New: func() any {
		return &Context{
			Buf:         make([]byte, 0, 1024),
			smallBuffer: make([]byte, 0, 20),
		}
	},
}

type Context struct {
	smallBuffer []byte // a small buffer to encode float and time.Time as string
	Buf         []byte
}

func newCtx() *Context {
	ctx := ctxPool.Get().(*Context)

	return ctx
}

func freeCtx(ctx *Context) {
	ctx.smallBuffer = ctx.smallBuffer[:0]
	ctx.Buf = ctx.Buf[:0]
	ctxPool.Put(ctx)
}
