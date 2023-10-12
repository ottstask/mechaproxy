package ingress

import (
	"fmt"
	"log"

	"github.com/valyala/fasthttp"
)

type HttpMessage struct {
	CallbackAddr string
	Payload      []byte
}

func writeIngressResponse(ctx *fasthttp.RequestCtx, status, code int, message string) {
	ctx.Response.SetStatusCode(status)
	ctx.Response.Header.Set("Content-Type", "application/json")
	fmt.Fprintf(ctx, `{"from": "ingress","code": %d, "message": "%s"}`, code, message)
}

func sendCallback(addr string, callRsp *fasthttp.Response) {
	if addr == "" {
		return
	}
	callbackReq := fasthttp.AcquireRequest()
	callbackReq.Header.SetMethod("POST")
	callbackReq.SetRequestURI(addr)
	callRsp.WriteTo(callbackReq.BodyWriter())

	callbackRsp := fasthttp.AcquireResponse()
	if err := fasthttp.Do(callbackReq, callbackRsp); err != nil {
		log.Println("callback error", err)
	}

	fasthttp.ReleaseRequest(callbackReq)
	fasthttp.ReleaseResponse(callbackRsp)
}
