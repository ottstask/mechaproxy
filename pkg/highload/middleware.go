package highload

import (
	"github.com/ottstack/mechaproxy/mechaproxy/middleware"
	"github.com/valyala/fasthttp"
)

func HighloadProtect() middleware.Middleware {
	return func(ctx *fasthttp.RequestCtx, next func(*fasthttp.RequestCtx) error) error {
		// check highload
		// while
		return next(ctx)
	}
}
