package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

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

	NatsUrl  string `name:"nats-url" env:"NATS_URL" default:"ns://127.0.0.1:4222" help:"NATS server url."`
	NatsJwt  string `name:"nats-jwt" env:"NATS_JWT"`
	NatsSeed string `name:"nats-seed" env:"NATS_SEED"`

	Cache cacheCmd `cmd:"" help:"Run a binary cache." default:"1"`
}

func buildLogger(opts logOptions) error {
	// configure logging
	var config zap.Config
	if opts.Development {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}

	// set log level
	l := opts.Level

	switch {
	case l == "debug":
		config.Level.SetLevel(zap.DebugLevel)
	case l == "info":
		config.Level.SetLevel(zap.InfoLevel)
	case l == "warn":
		config.Level.SetLevel(zap.WarnLevel)
	case l == "error":
		config.Level.SetLevel(zap.ErrorLevel)
	}

	var err error
	logger, err = config.Build()
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
