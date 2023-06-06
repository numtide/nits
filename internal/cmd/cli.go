package cmd

import (
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"

	log "github.com/inconshreveable/log15"
	"github.com/juju/errors"
	"github.com/numtide/nits/pkg/config"
)

type logOptions struct {
	Level string `enum:"debug,info,warn,error" env:"LOG_LEVEL" default:"warn" help:"Configure logging level."`
}

func (lo *logOptions) toLogger() log.Logger {
	l := log.New()
	level, err := log.LvlFromString(lo.Level)
	if err != nil {
		panic(err)
	}
	l.SetHandler(log.LvlFilterHandler(level, log.StreamHandler(os.Stderr, log.LogfmtFormat())))
	return l
}

type natsOptions struct {
	Url         string   `name:"url" env:"NATS_URL" default:"ns://127.0.0.1:4222" help:"NATS server url."`
	Jwt         string   `name:"jwt" env:"NATS_JWT"`
	JwtFile     *os.File `name:"jwt-file" env:"NATS_JWT_FILE"`
	Seed        string   `name:"seed" env:"NATS_SEED"`
	SeedFile    *os.File `name:"seed-file" env:"NATS_SEED_FILE"`
	HostKeyFile *os.File `name:"host-key-file" env:"NATS_HOST_KEY_FILE"`
}

func (n *natsOptions) toNatsConfig() (*config.Nats, error) {
	c := &config.Nats{
		Url:         n.Url,
		Jwt:         n.Jwt,
		Seed:        n.Seed,
		HostKeyFile: n.HostKeyFile,
	}

	if c.Seed == "" && n.SeedFile != nil {
		b, err := io.ReadAll(n.SeedFile)
		if err != nil {
			return nil, errors.Annotate(err, "failed to read nats seed file")
		}
		c.Seed = string(b)
	}

	if c.Jwt == "" && n.JwtFile != nil {
		b, err := io.ReadAll(n.JwtFile)
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

var Cli struct {
	Logging logOptions  `embed:"" prefix:"log-"`
	Nats    natsOptions `embed:"" prefix:"nats-"`

	Cache cacheCmd `cmd:"" help:"Binary cache."`
	Agent agentCmd `cmd:"" help:"Run an agent"`
}

func runCmd(main func(logger log.Logger, ctx context.Context) error) (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l := Cli.Logging.toLogger()

	go func() {
		l.Debug("listening for termination signals")
		c := make(chan os.Signal, 1) // we need to reserve to buffer size 1, so the notifier are not blocked
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		cancel()
	}()

	return main(l, ctx)
}
