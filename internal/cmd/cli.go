package cmd

import (
	"context"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/juju/errors"
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

func buildLogger(opts logOptions) error {
	// configure logging

	c := zap.Config{
		Encoding:    "json",
		Level:       zap.NewAtomicLevelAt(zapcore.InfoLevel),
		OutputPaths: []string{"stdout"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:   "message",
			LevelKey:     "level",
			EncodeLevel:  zapcore.CapitalLevelEncoder,
			TimeKey:      "time",
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}

	if opts.Development {
		c.Encoding = "console"
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
