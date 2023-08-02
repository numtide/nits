package keys

import (
	"github.com/nats-io/jwt/v2"
)

func Generate(dir string) (err error) {

	credsDir := dir + "/creds"
	signingDir := dir + "/signing"

	var op Set[jwt.OperatorClaims]
	var opSig Set[jwt.OperatorClaims]

	var sys Set[jwt.AccountClaims]
	var sysSig Set[jwt.AccountClaims]
	var sysUser Set[jwt.UserClaims]

	if op, err = NewOperatorSet(); err != nil {
		return
	} else if opSig, err = NewOperatorSet(); err != nil {
		return
	}

	op.Claims = jwt.NewOperatorClaims(op.PubKey)
	op.Claims.Name = "Nits"
	op.Claims.SigningKeys.Add(opSig.PubKey)
	if err = op.EncodeClaims(op.KP); err != nil {
		return
	}

	if err = op.WriteJwt(dir, "operator.jwt"); err != nil {
		return
	}
	if err = op.WriteCredentials(credsDir, "operator.creds"); err != nil {
		return
	}
	if err = opSig.WriteCredentials(signingDir, "operator.creds"); err != nil {
		return
	}

	// create system account
	if sys, err = NewAccountSet(); err != nil {
		return
	} else if sysSig, err = NewAccountSet(); err != nil {
		return
	}

	sys.Claims = jwt.NewAccountClaims(sys.PubKey)
	sys.Claims.Name = "SYS"
	sys.Claims.Issuer = opSig.PubKey
	sys.Claims.SigningKeys.Add(sysSig.PubKey)

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
		AccountTokenPosition: 3,
		Info: jwt.Info{
			Description: `Account specific monitoring stream`,
			InfoURL:     "https://docs.nats.io/nats-server/configuration/sys_accounts",
		},
	}}

	if err = sys.EncodeClaims(opSig.KP); err != nil {
		return
	}

	if err = sys.WriteJwt(dir, "sys.jwt"); err != nil {
		return
	}
	if err = sys.WriteCredentials(credsDir, "sys.creds"); err != nil {
		return
	}
	if err = sysSig.WriteCredentials(signingDir, "sys.creds"); err != nil {
		return
	}

	// create system user
	if sysUser, err = NewUserSet(); err != nil {
		return
	}

	sysUser.Claims = jwt.NewUserClaims(sysUser.PubKey)
	sysUser.Claims.Name = "sys"
	sysUser.Claims.IssuerAccount = sys.PubKey

	if err = sysUser.EncodeClaims(sysSig.KP); err != nil {
		return
	}

	return sysUser.WriteCredentials(credsDir, "sys-user.creds")
}
