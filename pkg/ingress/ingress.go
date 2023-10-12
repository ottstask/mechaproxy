package ingress

import (
	"fmt"
	"log"

	"github.com/kelseyhightower/envconfig"
	"github.com/ottstack/mechaproxy/pkg/innerip"
	"github.com/ottstask/configapi/pkg/meta"
	"github.com/ottstask/configapi/pkg/watchclient"
	"github.com/valyala/fasthttp"
)

const AsyncRequestHeader = "X-Async-Request"

var ingressConfig *serveConfig

type serveConfig struct {
	Addr string
}

func init() {
	ingressConfig = &serveConfig{
		Addr: ":17000",
	}

	err := envconfig.Process("ingress", ingressConfig)
	if err != nil {
		log.Fatal(err)
	}
}

func Serve() {
	log.Println("serving ingress", ingressConfig.Addr)
	if err := fasthttp.ListenAndServe(ingressConfig.Addr, handleIngress); err != nil {
		log.Fatal(err)
	}
}

func WatchConfig() {
	ingressConfigKey := meta.IngressKeyPrefix + innerip.Get()
	for val := range watchclient.Watch(ingressConfigKey, &meta.IngressConfig{}) {
		cfg := val.(*meta.IngressConfig)
		updateHostInfo(cfg.HostInfo)
	}
}

// handleIngress forword request to local service or send to queue (then wait for callback)
func handleIngress(ctx *fasthttp.RequestCtx) {
	host := string(ctx.Request.Header.Host())
	isAsyncRequest := string(ctx.Request.Header.Peek(AsyncRequestHeader)) == "1"
	// check limit, send to queue if full
	wrapper := getHostWrapper(host)
	if wrapper == nil {
		writeIngressResponse(ctx, 400, 404, fmt.Sprintf("Target %s not found", host))
		return
	}
	if !wrapper.checkLimit() {
		wrapper.queueRequest(ctx, isAsyncRequest)
		return
	}
	defer wrapper.releaseLimit()
	if isAsyncRequest {
		writeIngressResponse(ctx, 200, 0, "Received async request")
		go func() {
			if err := wrapper.forward(ctx); err != nil {
				log.Println("request async request error", err)
			}
		}()
		return
	}
	if err := wrapper.forward(ctx); err != nil {
		writeIngressResponse(ctx, 500, 502, fmt.Sprintf("Forward request to %s failed: %v", host, err))
	}

}
