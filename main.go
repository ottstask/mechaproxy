package main

import (
	"fmt"
	"os"

	"github.com/go-errors/errors"
	"github.com/ottstack/mechaproxy/internal/mechaproxy"
)

func main() {
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
