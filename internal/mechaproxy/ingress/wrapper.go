package ingress

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/go-errors/errors"
	"github.com/ottstack/mechaproxy/internal/mechaproxy/callback"
	"github.com/ottstack/mechaproxy/internal/mechaproxy/ingress/queueimp"
	"github.com/valyala/fasthttp"
)

var messageID uint64 = 1

type hostWrapper struct {
	limiter *concurrencyLimiter
	client  *fasthttp.HostClient
	queue   queueimp.MessageQueue
}

func newWraper(cfg *serveConfig) (*hostWrapper, error) {
	client := &fasthttp.HostClient{Addr: cfg.Target}
	limiter := newLimiter(cfg.ConcurrencyLimit)

	queue, err := queueimp.NewQueueImp(cfg.QueueSource)
	if err != nil {
		return nil, errors.New("new NewQueueImp error" + err.Error())
	}
	return &hostWrapper{client: client, limiter: limiter, queue: queue}, nil
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
