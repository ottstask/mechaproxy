package mechaproxy

import (
	"github.com/ottstack/mechaproxy/internal/mechaproxy/callback"
	"github.com/ottstack/mechaproxy/internal/mechaproxy/ingress"
)

func Start() error {
	if err := ensurePackage(); err != nil {
		return err
	}
	go startServer()
	go callback.Serve()
	return ingress.Serve()
}
