package exec

import (
	"os/exec"
)

func Nats(args ...string) *exec.Cmd {
	return exec.Command("nats", args...)
}
