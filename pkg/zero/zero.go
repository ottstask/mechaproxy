package zero

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/ottstack/mechaproxy/mechaproxy"
	"github.com/ottstack/mechaproxy/mechaproxy/middleware"
	"github.com/valyala/fasthttp"
)

func Zero() middleware.Middleware {
	var isActive int32 = 1
	var isStop int32 = 0
	var processLock = sync.Mutex{}
	inactiveInterval := time.Second * 30
	go func() {
		for range time.NewTicker(inactiveInterval).C {
			if atomic.LoadInt32(&isStop) == 0 && atomic.LoadInt32(&isActive) == 0 {
				atomic.StoreInt32(&isStop, 1)
				processLock.Lock()
				// double check with lock to prevent stop after start
				if atomic.LoadInt32(&isActive) == 0 {
					mechaproxy.StopServer()
				}
				processLock.Unlock()
			}
			atomic.StoreInt32(&isActive, 0)
		}
	}()
	return func(ctx *fasthttp.RequestCtx, next func(*fasthttp.RequestCtx) error) error {
		atomic.StoreInt32(&isActive, 1)
		if atomic.LoadInt32(&isStop) == 1 {
			processLock.Lock()
			mechaproxy.StartServer()
			processLock.Unlock()
			atomic.StoreInt32(&isStop, 0)
		}
		return next(ctx)
	}
}
