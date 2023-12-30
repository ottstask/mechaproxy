//go:build !unix

package mechaproxy

func reapProcess() {
}

func waitShutdown() <-chan bool {
	c := make(chan bool, 100)
	return c
}
