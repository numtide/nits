package cli

import (
	"fmt"
	"os"
	"syscall"

	"github.com/ztrue/shutdown"
)

type addClusterCmd struct {
	Name          string `help:"Name of the account under which Agents will run"`
	NitsPublicKey string `help:"Public key of the account under which Nits services run"`
}

func (c *addClusterCmd) Run() (err error) {
	// ensure shutdown hooks are run when the process exits
	go shutdown.Listen(syscall.SIGINT, syscall.SIGTERM)

	adminContext := fmt.Sprintf("%s-%s", c.Name, "Admin")

	var logsConfig, deploymentsConfig *os.File
	if logsConfig, err = openResourceLocally(streamConfig, "streams/agent-logs.json"); err != nil {
		return err
	}
	if deploymentsConfig, err = openResourceLocally(streamConfig, "streams/agent-deployments.json"); err != nil {
		return err
	}

	return cmdSequence(
		// default permissions is to deny all pubsub
		nsc("add", "account", "-n", c.Name, "--deny-pubsub", ">"),
		// enable Jetstream todo set sane default limits
		nsc(
			"edit", "account", "-n", c.Name,
			"--js-mem-storage", "-1",
			"--js-disk-storage", "-1",
			"--js-streams", "-1",
			"--js-consumer", "-1",
		),
		// import binary cache service
		nsc(
			"add", "import", "-a", c.Name,
			"-n", "binary-cache",
			"--service",
			"--src-account", c.NitsPublicKey,
			"--remote-subject", "NITS.CACHE.>",
			"--local-subject", "NITS.CACHE.>",
		),
		// create an admin user
		nsc("add", "user", "-a", c.Name, "-n", "Admin", "--allow-pubsub", ">"),
		// create a context for the admin user
		nsc("generate", "context", "-a", c.Name, "-u", "Admin", "--context", adminContext),
		// push the account changes to the NATS server
		nsc("push", "-a", c.Name),
		// create some streams
		nats("--context", adminContext, "stream", "add", "--config", logsConfig.Name()),
		nats("--context", adminContext, "stream", "add", "--config", deploymentsConfig.Name()),
	)
}
