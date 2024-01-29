package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	nsccmd "github.com/nats-io/nsc/v2/cmd"
	"github.com/numtide/nits/internal/cmd"

	"github.com/charmbracelet/log"
	nexec "github.com/numtide/nits/pkg/exec"
)

type addClusterCmd struct {
	Name string `arg:"" help:"Name of the account under which Agents will run"`
}

func (c *addClusterCmd) Run() (err error) {
	if err := Cmd.Log.ConfigureLog(); err != nil {
		return err
	}

	var op nsccmd.OperatorDescriber
	if op, err = nexec.DescribeOperator(); err != nil {
		return
	}

	log.Info("adding a new account", "name", c.Name)

	nsc := cmd.LogExec(nexec.Nsc("add", "account", "-n", c.Name, "--deny-pubsub", ">"))

	if _, err = nsc.Output(); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) && string(exit.Stderr) == fmt.Sprintf("Error: the account \"%s\" already exists\n", c.Name) {
			log.Warn("account already exists")
		} else {
			nexec.LogError("failed to add account", err)
			return
		}
	}

	log.Info("setting account permissions")

	// todo set sane default limits
	nsc = cmd.LogExec(nexec.Nsc("edit", "account", "-n", c.Name,
		"--js-mem-storage", "-1",
		"--js-disk-storage", "-1",
		"--js-streams", "-1",
		"--js-consumer", "-1",
	))

	if _, err = nsc.Output(); err != nil {
		nexec.LogError("failed to set account permissions", err)
		return
	}

	log.Info("creating an admin user", "name", "Admin")

	nsc = cmd.LogExec(nexec.Nsc("add", "user", "-a", c.Name, "-n", "Admin", "--allow-pubsub", ">"))

	if _, err = nsc.Output(); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) && string(exit.Stderr) == "Error: the user \"Admin\" already exists\n" {
			log.Warn("user already exists")
		} else {
			nexec.LogError("failed to add admin user", err)
			return
		}
	}

	adminContext := fmt.Sprintf("%s-%s-%s", op.Name, c.Name, "Admin")
	log.Info("generating an admin context", "name", adminContext)

	nsc = cmd.LogExec(nexec.Nsc("generate", "context", "-a", c.Name, "-u", "Admin", "--context", adminContext))

	if _, err = nsc.Output(); err != nil {
		nexec.LogError("failed to add an admin context", err)
		return
	}

	log.Info("pushing account to server", "name", c.Name)
	nsc = cmd.LogExec(nexec.Nsc("push", "-a", c.Name))

	if _, err = nsc.Output(); err != nil {
		nexec.LogError("failed to push account to server", err)
		return
	}

	var logsConfig, registryConfig *os.File
	if logsConfig, err = openResourceLocally(streamConfig, "streams/agent-logs.json"); err != nil {
		return err
	}
	if registryConfig, err = openResourceLocally(streamConfig, "streams/agent-registry.json"); err != nil {
		return err
	}

	log.Info("adding streams")

	nats := cmd.LogExec(nexec.Nats("--context", adminContext, "stream", "add", "--config", logsConfig.Name()))

	if _, err = nats.Output(); err != nil {
		nexec.LogError("failed to add logs stream", err)
		return
	}

	nats = cmd.LogExec(nexec.Nats("--context", adminContext, "stream", "add", "--config", registryConfig.Name()))

	if _, err = nats.Output(); err != nil {
		nexec.LogError("failed to add logs stream", err)
		return
	}

	log.Info("setup complete")

	return nil
}
