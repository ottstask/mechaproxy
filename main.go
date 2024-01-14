package main

import (
	"fmt"
	"os"

	"github.com/go-errors/errors"
	"github.com/ottstack/mechaproxy/mechaproxy"
	"github.com/ottstack/mechaproxy/pkg/highload"
	"github.com/ottstack/mechaproxy/pkg/recover"
	"github.com/ottstack/mechaproxy/pkg/zero"
)

func main() {
	mechaproxy.Use(recover.Recover)
	mechaproxy.Use(highload.HighloadProtect())

	mechaproxy.Use(zero.Zero())

	err := mechaproxy.Start()
	if err != nil {
		if e, ok := err.(*errors.Error); ok {
			fmt.Println("exit error:", e.ErrorStack())
			os.Exit(1)
		} else {
			panic(err)
		}
	}
}
