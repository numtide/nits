package agent

import (
	"github.com/numtide/nits/internal/cmd"
)

var Cmd struct {
	Nats    cmd.NatsOptions `embed:"" prefix:"nats-"`
	Logging cmd.LogOptions  `embed:"" prefix:"log-"`

	Run  runCmd  `cmd:"" help:"Run an agent." default:"1"`
	Nkey nkeyCmd `cmd:"" help:"Produce a User NKey from an ed25519 key"`
}
