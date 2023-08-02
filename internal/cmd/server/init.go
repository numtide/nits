package server

import (
	"github.com/charmbracelet/log"
	"github.com/numtide/nits/pkg/auth"
	"os"
)

const (
	DataDirPermissions = 0750
)

type initCmd struct {
}

func (r *initCmd) Run() (err error) {
	var l *log.Logger

	l, err = Cmd.Logging.ToLogger()
	if err != nil {
		return err
	}

	l = l.With("datadir", Cmd.DataDir)

	if _, err = os.Stat(Cmd.DataDir); os.IsNotExist(err) {
		l.Info("creating data directory")
		// create the data directory
		if err = os.MkdirAll(Cmd.DataDir, DataDirPermissions); err != nil {
			return err
		}
	} else {
		l.Info("data directory detected")
	}

	l.Info("generating NATS keys")
	// generate NATS keys
	// todo check if they already exist
	return auth.Generate(Cmd.DataDir)
}
