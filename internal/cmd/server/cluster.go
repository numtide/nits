package server

import (
	"github.com/charmbracelet/log"
	"github.com/numtide/nits/pkg/auth"
	"os"
)

type clusterCmd struct {
	Add addClusterCmd `cmd:""`
}

type addClusterCmd struct {
	Name string `arg:"" help:"A name for the new cluster"`
}

func (r *addClusterCmd) Run() (err error) {
	var l *log.Logger

	l, err = Cmd.Logging.ToLogger()
	if err != nil {
		return err
	}

	l = l.With("datadir", Cmd.DataDir)
	if _, err = os.Stat(Cmd.DataDir); os.IsNotExist(err) {
		return err
	}

	acct, err := auth.NewClusterAccount(r.Name, Cmd.DataDir)
	if err != nil {
		return
	}
	l.With("name", acct.Claims.Name).Info("Cluster account added")

	return
}
