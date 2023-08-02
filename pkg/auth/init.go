package auth

import (
	"fmt"
	"github.com/nats-io/jwt/v2"
)

func Generate(dir string) (err error) {

	var op Set[jwt.OperatorClaims]

	var sys Set[jwt.AccountClaims]
	var sysUser Set[jwt.UserClaims]

	var nits Set[jwt.AccountClaims]
	var nitsUser Set[jwt.UserClaims]

	if op, err = generateOperator(dir); err != nil {
		return
	} else if sys, err = generateSysAccount(dir, &op); err != nil {
		return
	} else if sysUser, err = generateSysUser(dir, &sys); err != nil {
		return
	} else if nits, err = generateNitsAccount(dir, &op); err != nil {
		return
	} else if nitsUser, err = generateNitsServerUser(dir, &nits); err != nil {
		return
	}

	println(fmt.Sprintf("Operator: %s", op.PubKey))
	println(fmt.Sprintf("Sys: %s", sys.PubKey))
	println(fmt.Sprintf("Sys User: %s", sysUser.PubKey))
	println(fmt.Sprintf("Nits: %s", nits.PubKey))
	println(fmt.Sprintf("Nits User: %s", nitsUser.PubKey))

	return
}

func generateOperator(dataDir string) (op Set[jwt.OperatorClaims], err error) {
	if op, err = NewOperatorSet(); err != nil {
		return
	}

	op.Claims = jwt.NewOperatorClaims(op.PubKey)
	op.Claims.Name = "Nits"
	if err = op.EncodeClaims(op.KP); err != nil {
		return
	}

	if err = op.WriteJwt(dataDir, "operator.jwt"); err != nil {
		return
	} else if err = op.WriteCredentials(dataDir+"/creds", "operator.creds"); err != nil {
		return
	}

	return
}

func generateSysAccount(dataDir string, op *Set[jwt.OperatorClaims]) (sys Set[jwt.AccountClaims], err error) {
	if sys, err = NewAccountSet(); err != nil {
		return
	}

	sys.Claims = jwt.NewAccountClaims(sys.PubKey)
	sys.Claims.Name = "SYS"
	sys.Claims.Issuer = op.PubKey

	sys.Claims.Exports = jwt.Exports{&jwt.Export{
		Name:                 "account-monitoring-services",
		Subject:              "$SYS.REQ.ACCOUNT.*.*",
		Type:                 jwt.Service,
		ResponseType:         jwt.ResponseTypeStream,
		AccountTokenPosition: 4,
		Info: jwt.Info{
			Description: `Request account specific monitoring services for: SUBSZ, CONNZ, LEAFZ, JSZ and INFO`,
			InfoURL:     "https://docs.nats.io/nats-server/configuration/sys_accounts",
		},
	}, &jwt.Export{
		Name:                 "account-monitoring-streams",
		Subject:              "$SYS.ACCOUNT.*.>",
		Type:                 jwt.Stream,
		ResponseType:         jwt.ResponseTypeStream,
		AccountTokenPosition: 3,
		Info: jwt.Info{
			Description: `Account specific monitoring stream`,
			InfoURL:     "https://docs.nats.io/nats-server/configuration/sys_accounts",
		},
	}}

	if err = sys.EncodeClaims(op.KP); err != nil {
		return
	} else if err = sys.WriteJwt(dataDir, "sys.jwt"); err != nil {
		return
	} else if err = sys.WriteJwt(dataDir+"/jwt", ""); err != nil {
		return
	} else if err = sys.WriteCredentials(fmt.Sprintf(""), "sys.creds"); err != nil {
		return
	}

	return
}

func generateSysUser(dataDir string, acct *Set[jwt.AccountClaims]) (user Set[jwt.UserClaims], err error) {
	if user, err = NewUserSet(); err != nil {
		return
	}

	user.Claims = jwt.NewUserClaims(user.PubKey)
	user.Claims.Name = "sys"
	user.Claims.IssuerAccount = acct.PubKey

	if err = user.EncodeClaims(acct.KP); err != nil {
		return
	} else if err = user.WriteCredentials(dataDir+"/creds", "sys-user.creds"); err != nil {
		return
	}

	return
}

func generateNitsAccount(dataDir string, op *Set[jwt.OperatorClaims]) (nits Set[jwt.AccountClaims], err error) {
	if nits, err = NewAccountSet(); err != nil {
		return
	}

	nits.Claims = jwt.NewAccountClaims(nits.PubKey)
	nits.Claims.Name = "Nits"
	nits.Claims.Issuer = op.PubKey

	nits.Claims.Limits.JetStreamLimits = jwt.JetStreamLimits{
		// unlimited usage of JetStream
		Streams:       -1,
		Consumer:      -1,
		DiskStorage:   -1,
		MemoryStorage: -1,
	}

	nits.Claims.Exports = jwt.Exports{
		&jwt.Export{
			Name:         "Nix Binary Cache",
			Subject:      "nits.cache",
			Type:         jwt.Service,
			ResponseType: jwt.ResponseTypeChunked,
		},
	}

	if err = nits.EncodeClaims(op.KP); err != nil {
		return
	} else if err = nits.WriteJwt(dataDir+"/jwt", ""); err != nil {
		return
	} else if err = nits.WriteCredentials(dataDir+"/creds", "nits.creds"); err != nil {
		return
	}

	return
}

func generateNitsServerUser(dataDir string, acct *Set[jwt.AccountClaims]) (user Set[jwt.UserClaims], err error) {
	if user, err = NewUserSet(); err != nil {
		return
	}

	user.Claims = jwt.NewUserClaims(user.PubKey)
	user.Claims.Name = "Nits Server"
	user.Claims.IssuerAccount = acct.PubKey

	if err = user.EncodeClaims(acct.KP); err != nil {
		return
	} else if err = user.WriteCredentials(dataDir+"/creds", "nits-server.creds"); err != nil {
		return
	}

	return
}
