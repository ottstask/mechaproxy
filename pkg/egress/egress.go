package egress

import (
	"fmt"
	"log"
	"strings"

	"github.com/valyala/fasthttp"
)

func Serve() {
	log.Println("serving egress", listenConfig.Addr)
	if err := fasthttp.ListenAndServe(listenConfig.Addr, handleEgress); err != nil {
		log.Fatal(err)
	}
}

// handleEgress find target namespace and deployment, forward request to one of downstream
func handleEgress(ctx *fasthttp.RequestCtx) {
	// find one downstream
	ipPort, newHost := matchDownStream(ctx)
	if ipPort == "" {
		writeEgressResponse(ctx, 500, 502, "no down stream")
		return
	}
	client := getHostClient(ipPort)
	// rename host to ip:port
	if newHost != "" {
		ctx.Request.SetHost(newHost)
	}
	// forward to ingress
	if err := client.Do(&ctx.Request, &ctx.Response); err != nil {
		msg := fmt.Sprint("forward ingress error: ", err)
		writeEgressResponse(ctx, 500, 502, msg)
	}
}

func getHostClient(addr string) *fasthttp.HostClient {
	// TODO: cache
	return &fasthttp.HostClient{Addr: addr}
}

func matchDownStream(ctx *fasthttp.RequestCtx) (target string, newHost string) {
	domain := string(ctx.Request.Host())
	srcIP := ctx.RemoteIP().String()

	// fill namespace to domain
	if !strings.ContainsRune(domain, '.') {
		ns := getNamespace(srcIP)
		domain = fmt.Sprintf("%s.%s", domain, ns)
	}

	// domain to ip
	if info := getDomainDownstream(domain); info != nil {
		return info.IngressAddr, info.Addr
	}
	return "", ""
}
