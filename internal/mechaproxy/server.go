package mechaproxy

import (
	"os"
	"os/exec"
)

func startServer() {
	cmd := exec.Command("sh", "start.sh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic("server error: " + err.Error())
	}
}
