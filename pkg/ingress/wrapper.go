package ingress

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ottstack/mechaproxy/pkg/callback"
	"github.com/ottstack/mechaproxy/pkg/ingress/queueimp"
	"github.com/ottstask/configapi/pkg/meta"
	"github.com/valyala/fasthttp"
)

var hostWrappers = sync.Map{}
var currConfig = atomic.Value{}
var messageID uint64 = 1

type hostWrapper struct {
	limiter *concurrencyLimiter
	client  *fasthttp.HostClient
	queue   queueimp.MessageQueue
}

func updateHostInfo(cfg map[string]*meta.IngressHostInfo) {
	lv := currConfig.Load()
	currCfg := map[string]*meta.IngressHostInfo{}
	if lv != nil {
		currCfg = lv.(map[string]*meta.IngressHostInfo)
	}
	// compare diff, add or replace
	for k, v := range cfg {
		if vv, ok := currCfg[k]; ok {
			val, _ := hostWrappers.Load(k)
			oldWrapper := val.(*hostWrapper)
			newWrapper := &hostWrapper{limiter: oldWrapper.limiter, client: oldWrapper.client}
			// check and update config
			hasUpdate := false
			if vv.Addr != v.Addr {
				oldWrapper.client.CloseIdleConnections()
				client := &fasthttp.HostClient{Addr: v.Addr}
				newWrapper.client = client
				hasUpdate = true
			}
			if vv.ConcurrencyLimit != v.ConcurrencyLimit {
				nl := newLimiter(v.ConcurrencyLimit)
				newWrapper.limiter = nl
			}
			if hasUpdate {
				hostWrappers.Store(k, newWrapper)
			}
			continue
		}
		client := &fasthttp.HostClient{Addr: v.Addr}
		lm := newLimiter(v.ConcurrencyLimit)

		queue, err := queueimp.NewQueueImp(v.QueueSource)
		if err != nil {
			log.Println("new NewQueueImp error", err)
			continue
		}

		newWrapper := &hostWrapper{limiter: lm, client: client, queue: queue}
		go newWrapper.consume()
		hostWrappers.Store(k, newWrapper)
	}

	// delete
	for k := range currCfg {
		if _, ok := cfg[k]; ok {
			continue
		}
		vv, _ := hostWrappers.Load(k)
		hc := vv.(*hostWrapper)
		hc.client.CloseIdleConnections()
		hostWrappers.Delete(k)
	}
	currConfig.Store(cfg)
}

func getHostWrapper(host string) *hostWrapper {
	val, ok := hostWrappers.Load(host)
	if !ok {
		return nil
	}
	return val.(*hostWrapper)
}

func (w *hostWrapper) checkLimit() bool {
	return w.limiter.acquire()
}

func (w *hostWrapper) releaseLimit() {
	w.limiter.release()
}

func (w *hostWrapper) forward(ctx *fasthttp.RequestCtx) error {
	return w.client.Do(&ctx.Request, &ctx.Response)
}

func (w *hostWrapper) queueRequest(ctx *fasthttp.RequestCtx, isAsync bool) {
	var msgID uint64 = atomic.AddUint64(&messageID, 1)
	// create channel before enqueue
	var waitChan *callback.WaitChan
	var callbackAddr string
	if !isAsync {
		ipPort := callback.GetAddress()
		callbackAddr = fmt.Sprintf("http://%s/%d", ipPort, msgID)
		waitChan = callback.NewCallbackChan(time.Minute, msgID, &ctx.Response)
	}

	header := ctx.Request.Header.Header()
	body := ctx.Request.Body()
	payload := bytes.NewBuffer(nil)
	payload.Write(header)
	payload.Write(body)

	msg := &queueimp.Message{Payload: payload.Bytes(), CallbackAddr: callbackAddr}

	w.queue.Push(msg)

	if isAsync {
		writeIngressResponse(ctx, 200, 0, "received")
		return
	} else {
		waitChan.Wait()
		callback.CallbackDone(msgID)
	}
}

func (w *hostWrapper) consume() {
	w.checkLimit()
	defer w.releaseLimit()
	for msg := range w.queue.Queue() {
		w.handleMessage(msg)
	}
}

func (w *hostWrapper) handleMessage(m *queueimp.Message) {
	req := &fasthttp.Request{}
	buf := bytes.NewBuffer(m.Payload)
	rd := bufio.NewReader(buf)
	if err := req.Read(rd); err != nil {
		log.Println("read request error", err)
		return
	}
	rsp := &fasthttp.Response{}
	if err := w.client.Do(req, rsp); err != nil {
		log.Println("client do error", err)
	} else {
		sendCallback(m.CallbackAddr, rsp)
	}

}
