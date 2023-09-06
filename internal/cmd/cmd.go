package cmd

import (
	"context"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/ztrue/shutdown"

	nsccmd "github.com/nats-io/nsc/v2/cmd"
	nexec "github.com/numtide/nits/pkg/exec"

	"github.com/numtide/nits/pkg/cache"

	"github.com/nix-community/go-nix/pkg/narinfo/signature"

	"github.com/charmbracelet/log"
	"github.com/numtide/nits/pkg/config"
)

type (
	Args     = []string
	ArgsList = []Args
)

type LogOptions struct {
	Level string `enum:"debug,info,warn,error,fatal" env:"LOG_LEVEL" default:"warn" help:"Configure logging level."`
}

func (lo *LogOptions) ConfigureLog() {
	log.SetReportTimestamp(true)
	log.SetTimeFormat(time.RFC3339)
	log.SetLevel(log.ParseLevel(lo.Level))
	log.SetTimeFormat(time.RFC3339)
}

type CacheOptions struct {
	Subject        string   `env:"NITS_CACHE_SUBJECT" default:"NITS.CACHE"`
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
			Priority:      o.Priority,
		},
	}, nil
}

func Run(main func(ctx context.Context) error) (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdown.Add(cancel)
	go shutdown.Listen(syscall.SIGINT, syscall.SIGTERM)

	return main(ctx)
}

type CacheProxyOptions struct {
	Subject   string `env:"NITS_CACHE_PROXY_SUBJECT" default:"NITS.CACHE"`
	PublicKey string `env:"NITS_CACHE_PROXY_PUBLIC_KEY"`
}

func (c *CacheProxyOptions) ToCacheProxyConfig() (*config.CacheProxy, error) {
	// todo validate format

	if c.PublicKey == "" {
		return nil, nil
	}

	return &config.CacheProxy{
		Subject:   c.Subject,
		PublicKey: c.PublicKey,
	}, nil
}

func DetectOperator() (operator nsccmd.OperatorDescriber, err error) {
	if operator, err = nexec.DescribeOperator(); err != nil {
		log.Error("failed to describe operator")
		return
	}

	log.Info("detected operator",
		"name", operator.Name,
		"serviceUrls", operator.OperatorServiceURLs,
		"accountServerUrl", operator.AccountServerURL,
	)
	return
}

func LogExec(cmd *exec.Cmd) *exec.Cmd {
	log.Debug(cmd.String())
	return cmd
}
