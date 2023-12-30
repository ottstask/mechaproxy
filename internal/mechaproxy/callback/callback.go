package callback

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/ottstack/mechaproxy/internal/utils/innerip"
	"github.com/valyala/fasthttp"
)

var callbackAddr string

type callbackConfig struct {
	Addr string
}

func Serve() error {
	cfg := &callbackConfig{
		Addr: ":18008",
	}
	err := envconfig.Process("CAllBACK", cfg)
	if err != nil {
		return err
	}
	callbackAddr = cfg.Addr
	if strings.HasPrefix(cfg.Addr, ":") || strings.HasPrefix(cfg.Addr, "0.0.0.0:") {
		port := cfg.Addr[strings.Index(cfg.Addr, ":"):]
		callbackAddr = innerip.Get() + port
	}
	log.Println("serving callback", callbackAddr)
	if err := fasthttp.ListenAndServe(cfg.Addr, handleCallback); err != nil {
		return err
	}
	return nil
}

func GetAddress() string {
	return callbackAddr
}

func handleCallback(ctx *fasthttp.RequestCtx) {
	path := ctx.Path()
	if len(path) <= 1 {
		return
	}
	id, _ := strconv.ParseUint(string(path[1:]), 10, 64)
	val, ok := rspMap.Load(id)
	if !ok {
		writeCallbackResponse(ctx, 400, 400, fmt.Sprintf("callback id %d not exists", id))
		return
	}
	originRsp := val.(*fasthttp.Response)
	buf := bytes.NewBuffer(ctx.Request.Body())
	rd := bufio.NewReader(buf)
	defer ctx.Request.ReleaseBody(0)
	if err := originRsp.Read(rd); err != nil {
		writeCallbackResponse(ctx, 500, 500, fmt.Sprintf("write response error %v", err))
	} else {
		writeCallbackResponse(ctx, 200, 0, "callback succ")
	}
	CallbackDone(id)
}
