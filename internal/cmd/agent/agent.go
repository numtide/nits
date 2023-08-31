package agent

import (
	"github.com/numtide/nits/pkg/nats"
)

var Cmd struct {
	Nats     nats.CliOptions `embed:"" prefix:"nats-"`
	LogLevel string          `enum:"debug,info,warn,error,fatal" env:"LOG_LEVEL" default:"warn" help:"Configure logging level."`

	Run  runCmd  `cmd:"" help:"Run an agent." default:"1"`
	Nkey nkeyCmd `cmd:"" help:"Produce a User NKey from an ed25519 key"`
}
