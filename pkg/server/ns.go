package server

import (
	"context"
	"github.com/charmbracelet/log"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/numtide/nits/pkg/auth"
	nitslog "github.com/numtide/nits/pkg/log"
	"time"
)

func (s *Server) runNats(ctx context.Context, log *log.Logger) (err error) {

	if err = auth.Generate(s.DataDir); err != nil {
		return err
	}

	var operator *auth.Set[jwt.OperatorClaims]
	var sysAcct *auth.Set[jwt.AccountClaims]

	if operator, err = auth.ReadOperatorJwt(s.DataDir + "/operator.jwt"); err != nil {
		return
	} else if sysAcct, err = auth.ReadAccountJwt(s.DataDir + "/sys.jwt"); err != nil {
		return
	}

	accResolver, err := server.NewDirAccResolver(
		s.DataDir+"/jwt",
		1024,
		5*time.Second,
		server.RenameDeleted,
	)

	opts := server.Options{
		JetStream:        true,
		StoreDir:         s.DataDir,
		SystemAccount:    sysAcct.PubKey,
		AccountResolver:  accResolver,
		TrustedOperators: []*jwt.OperatorClaims{operator.Claims},
	}

	s.srv, err = server.NewServer(&opts)
	if err != nil {
		return err
	}

	enable := s.srv.JetStreamEnabled()
	println(enable)

	go func() {
		<-ctx.Done()
		s.srv.Shutdown()
		s.srv.Shutdown()
	}()

	s.srv.SetLoggerV2(&nitslog.NatsLogAdapter{Logger: log}, false, false, false)
	s.srv.Start()

	return nil
}
