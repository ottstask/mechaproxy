package callback

import (
	"context"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

var rspMap = sync.Map{}
var cancalMap = sync.Map{}

type WaitChan struct {
	ch  chan struct{}
	id  uint64
	ctx context.Context
}

func NewCallbackChan(timeout time.Duration, id uint64, rsp *fasthttp.Response) *WaitChan {
	ch := make(chan struct{})
	var canf context.CancelFunc

	ctx, canf := context.WithTimeout(context.Background(), timeout)
	rspMap.Store(id, rsp)
	cancalMap.Store(id, canf)

	wc := &WaitChan{id: id, ch: ch, ctx: ctx}
	return wc
}

func (w *WaitChan) Wait() {
	<-w.ctx.Done()
}

func CallbackDone(id uint64) {
	val, ok := cancalMap.Load(id)
	if ok {
		canf := val.(context.CancelFunc)
		canf()
		cancalMap.Delete(id)
		rspMap.Delete(id)
	}
}
