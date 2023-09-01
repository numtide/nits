package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/charmbracelet/log"
	nsccmd "github.com/nats-io/nsc/v2/cmd"

	nexec "github.com/numtide/nits/pkg/exec"

	"github.com/ztrue/shutdown"
)

type addClusterCmd struct {
	Name string `help:"Name of the account under which Agents will run"`
}

func (c *addClusterCmd) Run() (err error) {
	// ensure shutdown hooks are run when the process exits
	go shutdown.Listen(syscall.SIGINT, syscall.SIGTERM)

	var operator nsccmd.OperatorDescriber
	if operator, err = nexec.DescribeOperator(); err != nil {
		log.Error("failed to describe operator")
		return
	}

	log.Info("detected operator",
		"name", operator.Name,
		"serviceUrls", operator.OperatorServiceURLs,
		"accountServerUrl", operator.AccountServerURL,
	)

	log.Info("adding a new account", "name", c.Name)

	cmd := nexec.Nsc("add", "account", "-n", c.Name, "--deny-pubsub", ">")
	log.Debug(cmd.String())

	if _, err = cmd.Output(); err != nil {
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
	cmd = nexec.Nsc("edit", "account", "-n", c.Name,
		"--js-mem-storage", "-1",
		"--js-disk-storage", "-1",
		"--js-streams", "-1",
		"--js-consumer", "-1",
	)
	log.Debug(cmd.String())

	if _, err = cmd.Output(); err != nil {
		nexec.LogError("failed to set account permissions", err)
		return
	}

	log.Info("creating an admin user", "name", "Admin")

	cmd = nexec.Nsc("add", "user", "-a", c.Name, "-n", "Admin", "--allow-pubsub", ">")
	log.Debug(cmd.String())

	if _, err = cmd.Output(); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) && string(exit.Stderr) == "Error: the user \"Admin\" already exists\n" {
			log.Warn("user already exists")
		} else {
			nexec.LogError("failed to add admin user", err)
			return
		}
	}

	adminContext := fmt.Sprintf("%s-%s", c.Name, "Admin")
	log.Info("generating an admin context", "name", adminContext)

	cmd = nexec.Nsc("generate", "context", "-a", c.Name, "-u", "Admin", "--context", adminContext)
	log.Debug(cmd.String())

	if _, err = cmd.Output(); err != nil {
		nexec.LogError("failed to add an admin context", err)
		return
	}

	log.Info("pushing account to server", "name", c.Name)
	cmd = nexec.Nsc("push", "-a", c.Name)
	log.Debug(cmd.String())

	if _, err = cmd.Output(); err != nil {
		nexec.LogError("failed to push account to server", err)
		return
	}

	var logsConfig *os.File
	if logsConfig, err = openResourceLocally(streamConfig, "streams/agent-logs.json"); err != nil {
		return err
	}

	log.Info("adding streams")

	cmd = nexec.Nats("--context", adminContext, "stream", "add", "--config", logsConfig.Name())
	log.Debug(cmd.String())

	if _, err = cmd.Output(); err != nil {
		nexec.LogError("failed to add logs stream", err)
		return
	}

	log.Info("setup complete")

	return nil
}
