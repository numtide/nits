package cmd

import (
	"context"
	"github.com/nix-community/go-nix/pkg/narinfo/signature"
	"github.com/numtide/nits/pkg/services/cache"
	"io"
	"os"
	"os/signal"
	"syscall"

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
	InboxFormat string   `name:"inbox-format" env:"NATS_INBOX_FORMAT"`
}

func (n *NatsOptions) ToNatsConfig() (*config.Nats, error) {
	c := &config.Nats{
		Url:         n.Url,
		Jwt:         n.Jwt,
		Seed:        n.Seed,
		HostKeyFile: n.HostKeyFile,
		InboxFormat: n.InboxFormat,
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
	Subject        string   `env:"NITS_CACHE_SUBJECT" default:"nits.cache"`
	Group          string   `env:"NITS_CACHE_GROUP" default:"cache"`
	StoreDir       string   `env:"NITS_CACHE_STORE_DIR" default:"/nix/store"`
	WantMassQuery  bool     `env:"NITS_CACHE_WANT_MASS_QUERY" default:"true"`
	Priority       int      `env:"NITS_CACHE_PRIORITY" default:"1"`
	PrivateKeyFile *os.File `env:"NITS_CACHE_PRIVATE_KEY_FILE" required:""`
}

func (o *CacheOptions) ToCacheOptions() (*cache.Options, error) {

	bytes, err := io.ReadAll(o.PrivateKeyFile)
	if err != nil {
		return nil, err
	}

	secretKey, err := signature.LoadSecretKey(string(bytes))
	if err != nil {
		return nil, err
	}

	return &cache.Options{
		SecretKey: &secretKey,
		Subject:   o.Subject,
		Group:     o.Group,
		Info: &cache.Info{
			StoreDir:      o.StoreDir,
			WantMassQuery: o.WantMassQuery,
			Priority:      o.Priority},
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
