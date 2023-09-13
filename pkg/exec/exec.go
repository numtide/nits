package exec

import (
	"os/exec"

	"github.com/charmbracelet/log"
	"github.com/juju/errors"
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
