package cli

import (
	"io"
	"os"
	"os/exec"
)

var Cmd struct {
	Add struct {
		Cache   addCacheCmd   `cmd:""`
		Cluster addClusterCmd `cmd:""`
		Agent   addAgentCmd   `cmd:""`
	} `cmd:"" help:"Add assets such as clusters and agents"`

	Deploy deployCmd `cmd:"" help:"deploy a NixOS closure to one or more agents"`
}

type shellCmd struct {
	name   string
	args   []string
	stdout io.Writer
	stderr io.Writer
}

func (sc *shellCmd) Exec() (err error) {
	c := exec.Command(sc.name, sc.args...)
	c.Stdout = sc.stdout
	c.Stderr = sc.stderr

	if c.Stdout == nil {
		c.Stdout = os.Stdout
	}
	if c.Stderr == nil {
		c.Stderr = os.Stderr
	}

	return c.Run()
}

func cmdSequence(cmds ...shellCmd) (err error) {
	for _, c := range cmds {
		if err = c.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func nsc(args ...string) shellCmd {
	return shellCmd{name: "nsc", args: args}
}

func nats(args ...string) shellCmd {
	return shellCmd{name: "nats", args: args}
}
