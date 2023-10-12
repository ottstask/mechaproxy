package egress

import (
	"fmt"

	"github.com/valyala/fasthttp"
)

func writeEgressResponse(ctx *fasthttp.RequestCtx, status, code int, message string) {
	ctx.Response.SetStatusCode(status)
	ctx.Response.Header.Set("Content-Type", "application/json")
	fmt.Fprintf(ctx, `{"from": "egress","code": %d, "message": "%s"}`, code, message)
}
