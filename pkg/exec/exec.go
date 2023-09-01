package exec

import (
	"github.com/charmbracelet/log"
	"github.com/juju/errors"
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

func LogError(msg string, err error) {
	var exit *exec.ExitError
	var kv []interface{}
	if errors.As(err, &exit) {
		kv = append(kv, "exitCode", exit.ExitCode(), "error", string(exit.Stderr))
	} else {
		kv = append(kv, "error", err)
	}
	log.Error(msg, kv...)
}
