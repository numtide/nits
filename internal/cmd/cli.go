package cmd

import (
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/juju/errors"
	"github.com/nats-io/nkeys"
	"github.com/numtide/nits/pkg/config"

	"github.com/ethereum/go-ethereum/log"
	"go.uber.org/zap"
)

var logger *zap.Logger

type logOptions struct {
	Level       string `enum:"debug,info,warn,error" env:"LOG_LEVEL" default:"warn" help:"Configure logging level."`
	Development bool   `env:"LOG_DEVELOPMENT" default:"false" help:"Configure development style log output."`
}

var Cli struct {
	Logging logOptions `embed:"" prefix:"log-"`

	NatsUrl             string   `name:"nats-url" env:"NATS_URL" default:"ns://127.0.0.1:4222" help:"NATS server url."`
	NatsJwt             string   `name:"nats-jwt" env:"NATS_JWT"`
	NatsSeed            string   `name:"nats-seed" env:"NATS_SEED"`
	NatsHostKeyFile     *os.File `nmae:"nats-host-key-file" env:"NATS_HOST_KEY_FILE"`
	NatsCredentialsFile *os.File `name:"nats-credentials-file" env:"NATS_CREDENTIALS_FILE"`

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

	if Cli.NatsCredentialsFile != nil {
		bytes, err := io.ReadAll(Cli.NatsCredentialsFile)
		if err != nil {
			return nil, errors.Annotate(err, "failed to read nats credentials file")
		}

		jwt, err := nkeys.ParseDecoratedJWT(bytes)
		if err != nil {
			return nil, err
		}

		keyPair, err := nkeys.ParseDecoratedNKey(bytes)
		if err != nil {
			return nil, err
		}

		seed, err := keyPair.Seed()
		if err != nil {
			return nil, err
		}

		c.Jwt = jwt
		c.Seed = string(seed)
	}

	if c.Jwt == "" {
		return nil, errors.New("nats jwt cannot be empty")
	}

	if c.Seed == "" && c.HostKeyFile == nil {
		return nil, errors.New("one of nats seed or nats host key must be set")
	}

	return c, nil
}

func buildLogger(opts logOptions) error {
	// configure logging
	var c zap.Config
	if opts.Development {
		c = zap.NewDevelopmentConfig()
	} else {
		c = zap.NewProductionConfig()
	}

	// set log level
	l := opts.Level

	switch {
	case l == "debug":
		c.Level.SetLevel(zap.DebugLevel)
	case l == "info":
		c.Level.SetLevel(zap.InfoLevel)
	case l == "warn":
		c.Level.SetLevel(zap.WarnLevel)
	case l == "error":
		c.Level.SetLevel(zap.ErrorLevel)
	}

	var err error
	logger, err = c.Build()
	return err
}

func runCmd(main func(ctx context.Context) error) (err error) {
	if err = buildLogger(Cli.Logging); err != nil {
		return err
	}

	defer func(Logger *zap.Logger) {
		_ = Logger.Sync()
	}(logger)

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
