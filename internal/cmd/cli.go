package cmd

import (
	"context"
	log "github.com/inconshreveable/log15"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/juju/errors"
	"github.com/numtide/nits/pkg/config"
)

type logOptions struct {
	Level string `enum:"debug,info,warn,error" env:"LOG_LEVEL" default:"warn" help:"Configure logging level."`
}

var Cli struct {
	Logging logOptions `embed:"" prefix:"log-"`

	NatsUrl         string   `name:"nats-url" env:"NATS_URL" default:"ns://127.0.0.1:4222" help:"NATS server url."`
	NatsJwt         string   `name:"nats-jwt" env:"NATS_JWT"`
	NatsJwtFile     *os.File `name:"nats-jwt-file" env:"NATS_JWT_FILE"`
	NatsSeed        string   `name:"nats-seed" env:"NATS_SEED"`
	NatsSeedFile    *os.File `name:"nats-seed-file" env:"NATS_SEED_FILE"`
	NatsHostKeyFile *os.File `name:"nats-host-key-file" env:"NATS_HOST_KEY_FILE"`

	Cache cacheCmd `cmd:"" help:"Binary cache."`
	Agent agentCmd `cmd:"" help:"Run an agent"`
}

func natsConfig() (*config.Nats, error) {
	c := &config.Nats{
		Url:         Cli.NatsUrl,
		Jwt:         Cli.NatsJwt,
		Seed:        Cli.NatsSeed,
		HostKeyFile: Cli.NatsHostKeyFile,
	}

	if c.Seed == "" && Cli.NatsSeedFile != nil {
		b, err := io.ReadAll(Cli.NatsSeedFile)
		if err != nil {
			return nil, errors.Annotate(err, "failed to read nats seed file")
		}
		c.Seed = string(b)
	}

	if c.Jwt == "" && Cli.NatsJwtFile != nil {
		b, err := io.ReadAll(Cli.NatsJwtFile)
		if err != nil {
			return nil, errors.Annotate(err, "failed to read nats jwt file")
		}
		c.Jwt = string(b)
	}

	if c.Seed == "" && c.HostKeyFile == nil {
		return nil, errors.New("one of nats seed or nats host key must be set")
	}

	return c, nil
}

func runCmd(main func(ctx context.Context) error) (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		log.Debug("listening for termination signals")
		c := make(chan os.Signal, 1) // we need to reserve to buffer size 1, so the notifier are not blocked
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		cancel()
	}()
	return main(ctx)
}
