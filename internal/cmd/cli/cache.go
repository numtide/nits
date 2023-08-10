package cli

import (
	"fmt"

	"github.com/nats-io/jwt/v2"
)

type addCacheCmd struct {
	Account string `default:"Nits" help:"Name of the account under which the cache service will run"`
}

func (c *addCacheCmd) Run() (err error) {
	cacheContext := fmt.Sprintf("%s-%s", c.Account, "Cache")

	return cmdSequence(
		// enable Jetstream for
		// todo set sane default limits
		nsc(
			"edit", "account", "-n", c.Account,
			"--js-mem-storage", "-1",
			"--js-disk-storage", "-1",
			"--js-streams", "-1",
			"--js-consumer", "-1",
		),
		// export binary cache service
		nsc(
			"add", "export", "-a", c.Account, "--service",
			"--name", "binary-cache",
			"--subject", "NITS.CACHE.>",
			"--response-type", jwt.ResponseTypeStream,
		),
		// generate a user for the cache service
		nsc(
			"add", "user", "-a", c.Account, "--name", "Cache",
		),
		// create a context for the admin user
		nsc("generate", "context", "-a", c.Account, "-u", "Cache", "--context", cacheContext),
		// push updated account jwt
		nsc("push", "-a", c.Account),
	)
}
