package encoder

import (
	"sync"
)

var ctxPool = sync.Pool{
	New: func() any {
		return &Ctx{
			Buf:         make([]byte, 0, 1024),
			smallBuffer: make([]byte, 0, 20),
		}
	},
}

type Ctx struct {
	smallBuffer []byte // a small buffer to encode float and time.Time as string
	Buf         []byte
}

func newCtx() *Ctx {
	ctx := ctxPool.Get().(*Ctx)

	return ctx
}

func freeCtx(ctx *Ctx) {
	ctx.smallBuffer = ctx.smallBuffer[:0]
	ctx.Buf = ctx.Buf[:0]
	ctxPool.Put(ctx)
}
