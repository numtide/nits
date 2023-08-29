package exec

import (
	"io"
	"os/exec"
)

type Command func() error

func Sequence(commands ...*exec.Cmd) (err error) {
	for _, cmd := range commands {
		if err = cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func WithOutput(cmd *exec.Cmd, stdout io.Writer, stderr io.Writer) *exec.Cmd {
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd
}
