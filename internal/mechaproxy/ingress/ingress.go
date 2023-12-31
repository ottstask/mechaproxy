package ingress

import (
	"fmt"
	"log"

	"github.com/kelseyhightower/envconfig"
	"github.com/valyala/fasthttp"
)

const enqueueRequestHeader = "X-Enqueue-Request"

var proxyConfig *serveConfig
var wrapper *hostWrapper

type serveConfig struct {
	Addr             string
	Target           string
	ConcurrencyLimit int
	QueueSource      string
}

func init() {
	proxyConfig = &serveConfig{
		Addr:             ":8080",
		Target:           "127.0.0.1:8081",
		ConcurrencyLimit: 1000,
		QueueSource:      "local://",
	}

	err := envconfig.Process("INGRESS", proxyConfig)
	if err != nil {
		log.Fatal(err)
	}
}

func Serve() (err error) {
	wrapper, err = newWraper(proxyConfig)
	if err != nil {
		return err
	}
	go wrapper.consume()
	log.Println("serving ingress", proxyConfig.Addr)
	if err := fasthttp.ListenAndServe(proxyConfig.Addr, handleIngress); err != nil {
		return err
	}
	return nil
}

// handleIngress forword request to local service or send to queue (then wait for callback)
func handleIngress(ctx *fasthttp.RequestCtx) {
	forceEnqueue := string(ctx.Request.Header.Peek(enqueueRequestHeader)) == "1"
	if forceEnqueue || !wrapper.checkLimit() {
		wrapper.queueRequest(ctx, forceEnqueue)
		return
	}
	defer wrapper.releaseLimit()
	if err := wrapper.forward(ctx); err != nil {
		writeIngressResponse(ctx, 500, 502, fmt.Sprintf("Forward request to %s failed: %v", proxyConfig.Target, err))
	}
}
