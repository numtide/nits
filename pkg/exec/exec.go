package exec

import (
	"io"
	"os"
	"os/exec"
)

type ShellCmd struct {
	Name   string
	Args   []string
	Stdout io.Writer
	Stderr io.Writer
}

func (sc *ShellCmd) Exec() (err error) {
	c := exec.Command(sc.Name, sc.Args...)
	c.Stdout = sc.Stdout
	c.Stderr = sc.Stderr

	if c.Stdout == nil {
		c.Stdout = os.Stdout
	}
	if c.Stderr == nil {
		c.Stderr = os.Stderr
	}

	return c.Run()
}

func CmdSequence(cmds ...ShellCmd) (err error) {
	for _, c := range cmds {
		if err = c.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func Nsc(args ...string) ShellCmd {
	return ShellCmd{Name: "nsc", Args: args}
}

func Nats(args ...string) ShellCmd {
	return ShellCmd{Name: "nats", Args: args}
}
