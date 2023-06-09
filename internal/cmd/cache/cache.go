package cache

import (
	"os"

	"github.com/numtide/nits/internal/cmd"
	"github.com/numtide/nits/pkg/cache"
)

type Options struct {
	StoreDir       string   `env:"NITS_CACHE_STORE_DIR" default:"/nix/store"`
	WantMassQuery  bool     `env:"NITS_CACHE_WANT_MASS_QUERY" default:"true"`
	Priority       int      `env:"NITS_CACHE_PRIORITY" default:"1"`
	PrivateKeyFile *os.File `env:"NITS_CACHE_PRIVATE_KEY_FILE"`
	BindAddress    string   `env:"NITS_CACHE_BIND_ADDRESS" default:"localhost:3000"`
}

func (o *Options) ToCacheOptions() ([]cache.Option, error) {
	return []cache.Option{
		cache.BindAddress(o.BindAddress),
		cache.InfoConfig(o.StoreDir, o.WantMassQuery, o.Priority),
		cache.PrivateKeyFile(o.PrivateKeyFile),
	}, nil
}

var Cmd struct {
	Nats    cmd.NatsOptions `embed:"" prefix:"nats-"`
	Logging cmd.LogOptions  `embed:"" prefix:"log-"`

	Cache Options `embed:"" prefix:""`

	Run runCmd `cmd:"" help:"Run a binary cache."`
	GC  gcCmd  `cmd:"" help:"Garbage collect the binary cache."`
}
