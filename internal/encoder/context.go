package encoder

import (
	"sync"
	"unsafe"
)

var ctxPool = sync.Pool{
	New: func() any {
		return &Context{
			ptrSeen: make(map[unsafe.Pointer]struct{}, 100),
			Buf:     make([]byte, 0, 16*1024),
		}
	},
}

type empty = struct{}

type Context struct {
	ptrLevel int
	ptrSeen  map[unsafe.Pointer]empty
	Buf      []byte
}

func NewCtx() *Context {
	return ctxPool.Get().(*Context)
}

func FreeCtx(ctx *Context) {
	if cap(ctx.Buf) >= 100*1024*1024 { // drop buffer that are too long
		return
	}

	ctx.ptrLevel = 0
	clear(ctx.ptrSeen)
	ctx.Buf = ctx.Buf[:0]
	ctxPool.Put(ctx)
}
