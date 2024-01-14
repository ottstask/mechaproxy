package mechaproxy

import (
	"fmt"

	"github.com/go-errors/errors"
	"github.com/kelseyhightower/envconfig"
	"github.com/ottstack/mechaproxy/mechaproxy/middleware"
	"github.com/valyala/fasthttp"
)

var globalClient *fasthttp.HostClient
var middlewares []middleware.Middleware

type serveConfig struct {
	Addr   string
	Target string
}

func Start() error {
	go reapProcess()
	StartServer()
	return serveProxy()
}

func Use(m middleware.Middleware) {
	middlewares = append(middlewares, m)
}

func serveProxy() error {
	cfg := &serveConfig{
		Addr:   ":8080",
		Target: "127.0.0.1:8081",
	}
	err := envconfig.Process("INGRESS", cfg)
	if err != nil {
		return errors.New(err)
	}
	globalClient = &fasthttp.HostClient{Addr: cfg.Target}
	if err := fasthttp.ListenAndServe(cfg.Addr, HandleRequest); err != nil {
		return err
	}
	return nil
}

func HandleRequest(ctx *fasthttp.RequestCtx) {
	next := handleRequest
	for i := range middlewares {
		mware := middlewares[len(middlewares)-i-1]
		next = func(mm func(ctx *fasthttp.RequestCtx) error) func(ctx *fasthttp.RequestCtx) error {
			return func(ctx *fasthttp.RequestCtx) error {
				return mware(ctx, mm)
			}
		}(next)
	}
	if err := next(ctx); err != nil {
		ctx.Response.SetStatusCode(500)
		ctx.Response.Header.Set("Content-Type", "application/json")
		fmt.Fprintf(ctx, `{"from": "ingress","code": %d, "message": "%s"}`, 500, err.Error())
	}
}

func handleRequest(ctx *fasthttp.RequestCtx) error {
	return globalClient.Do(&ctx.Request, &ctx.Response)
}
