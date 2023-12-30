package callback

import (
	"fmt"

	"github.com/valyala/fasthttp"
)

func writeCallbackResponse(ctx *fasthttp.RequestCtx, status, code int, message string) {
	ctx.Response.SetStatusCode(status)
	ctx.Response.Header.Set("Content-Type", "application/json")
	fmt.Fprintf(ctx, `{"from": "callback","code": %d, "message": "%s"}`, code, message)
}
