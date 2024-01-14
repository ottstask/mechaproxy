package middleware

import "github.com/valyala/fasthttp"

type Middleware func(ctx *fasthttp.RequestCtx, next func(*fasthttp.RequestCtx) error) error
