package mechaproxy

import (
	"os"
	"os/exec"
)

func StartServer() {
	args := os.Args
	if len(args) <= 1 {
		return
	}
	cmd := exec.Command(args[1], args[2:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// TODO: check process ok and
	if err := cmd.Run(); err != nil {
		panic("server error: " + err.Error())
	}
}

func StopServer() {
	// check recent start

	// kill process group
}
