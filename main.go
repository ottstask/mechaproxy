package main

import (
	"github.com/ottstack/mechaproxy/pkg/callback"
	"github.com/ottstack/mechaproxy/pkg/egress"
	"github.com/ottstack/mechaproxy/pkg/ingress"
	"github.com/ottstask/configapi/pkg/watchclient"
)

func main() {
	watchclient.InitWatcher()

	// ingress: forword request to local container or send to queue (then wait for callback)
	// consumer: consume request and send it to local container, callback to producer if need
	// callback: find origin request and send response
	go ingress.WatchConfig()
	go ingress.Serve()
	go callback.Serve()

	// egress: find target namespace and deployment, forward request to one of downstream
	go egress.WatchDomain()
	go egress.WatchConfig()
	egress.Serve()
}
