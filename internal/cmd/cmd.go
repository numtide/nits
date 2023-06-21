package cmd

import (
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/numtide/nits/pkg/cache"

	log "github.com/inconshreveable/log15"
	"github.com/juju/errors"
	"github.com/numtide/nits/pkg/config"
)

type LogOptions struct {
	Level string `enum:"debug,info,warn,error" env:"LOG_LEVEL" default:"warn" help:"Configure logging level."`
}

func (lo *LogOptions) ToLogger() log.Logger {
	l := log.New()
	level, err := log.LvlFromString(lo.Level)
	if err != nil {
		panic(err)
	}
	l.SetHandler(log.LvlFilterHandler(level, log.StreamHandler(os.Stderr, log.LogfmtFormat())))
	return l
}

type NatsOptions struct {
	Url         string   `name:"url" env:"NATS_URL" default:"ns://127.0.0.1:4222" help:"NATS server url."`
	Jwt         string   `name:"jwt" env:"NATS_JWT"`
	JwtFile     *os.File `name:"jwt-file" env:"NATS_JWT_FILE"`
	Seed        string   `name:"seed" env:"NATS_SEED"`
	SeedFile    *os.File `name:"seed-file" env:"NATS_SEED_FILE"`
	HostKeyFile *os.File `name:"host-key-file" env:"NATS_HOST_KEY_FILE"`
	InboxPrefix string   `name:"inbox-prefix" env:"NATS_INBOX_PREFIX"`
}

func (n *NatsOptions) ToNatsConfig() (*config.Nats, error) {
	c := &config.Nats{
		Url:         n.Url,
		Jwt:         n.Jwt,
		Seed:        n.Seed,
		HostKeyFile: n.HostKeyFile,
		InboxPrefix: n.InboxPrefix,
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

type CacheOptions struct {
	StoreDir       string   `env:"NITS_CACHE_STORE_DIR" default:"/nix/store"`
	WantMassQuery  bool     `env:"NITS_CACHE_WANT_MASS_QUERY" default:"true"`
	Priority       int      `env:"NITS_CACHE_PRIORITY" default:"1"`
	PrivateKeyFile *os.File `env:"NITS_CACHE_PRIVATE_KEY_FILE"`
	BindAddress    string   `env:"NITS_CACHE_BIND_ADDRESS" default:"localhost:3000"`
}

func (o *CacheOptions) ToCacheOptions() ([]cache.Option, error) {
	return []cache.Option{
		cache.BindAddress(o.BindAddress),
		cache.InfoConfig(o.StoreDir, o.WantMassQuery, o.Priority),
		cache.PrivateKeyFile(o.PrivateKeyFile),
	}, nil
}

func Run(logger log.Logger, main func(ctx context.Context) error) (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		logger.Debug("listening for termination signals")
		c := make(chan os.Signal, 1) // we need to reserve to buffer size 1, so the notifier are not blocked
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		cancel()
	}()

	return main(ctx)
}
