package server

import (
	"github.com/numtide/nits/internal/cmd"
)

var Cmd struct {
	Logging cmd.LogOptions `embed:"" prefix:"log-"`

	DataDir string `env:"NATS_SERVER_DATA_DIR" default:"./datadir"`

	Run     runCmd     `cmd:"" default:"1"`
	Init    initCmd    `cmd:""`
	Cluster clusterCmd `cmd:""`
}
