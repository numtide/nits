package exec

import (
	"os"
	"os/exec"
)

func Nsc(args ...string) *exec.Cmd {
	return WithOutput(exec.Command("nsc", args...), os.Stdout, os.Stderr)
}

func Nats(args ...string) *exec.Cmd {
	return WithOutput(exec.Command("nats", args...), os.Stdout, os.Stderr)
}
