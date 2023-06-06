package cache

import (
	"github.com/numtide/nits/internal/cmd"
	"github.com/numtide/nits/pkg/cache"
	"os"
)

var Cmd struct {
	Nats    cmd.NatsOptions `embed:"" prefix:"nats-"`
	Logging cmd.LogOptions  `embed:"" prefix:"log-"`

	StoreDir       string   `env:"NITS_CACHE_STORE_DIR" default:"/nix/store"`
	WantMassQuery  bool     `env:"NITS_CACHE_WANT_MASS_QUERY" default:"true"`
	Priority       int      `env:"NITS_CACHE_PRIORITY" default:"1"`
	PrivateKeyFile *os.File `env:"NITS_CACHE_PRIVATE_KEY_FILE"`

	Run runCmd `cmd:"" help:"Run a binary cache."`
	GC  gcCmd  `cmd:"" help:"Garbage collect the binary cache."`
}

func cacheOptions() ([]cache.Option, error) {
	nats, err := Cmd.Nats.ToNatsConfig()
	if err != nil {
		return nil, err
	}

	return []cache.Option{
		cache.NatsConfig(nats),
		cache.InfoConfig(Cmd.StoreDir, Cmd.WantMassQuery, Cmd.Priority),
		cache.PrivateKeyFile(Cmd.PrivateKeyFile),
	}, nil
}
