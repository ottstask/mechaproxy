package recover

import (
	"fmt"
	"log"
	"runtime"

	"github.com/valyala/fasthttp"
)

func Recover(ctx *fasthttp.RequestCtx, next func(*fasthttp.RequestCtx) error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
			buf := make([]byte, 5*1024)
			n := runtime.Stack(buf, false)
			if n < len(buf) {
				buf = buf[:n]
			} else {
				buf = append(buf, []byte("...")...)
			}
			log.Printf("panic: %v\n %s", r, string(buf))
		}
	}()
	return next(ctx)
}
