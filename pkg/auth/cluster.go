package auth

import "github.com/nats-io/jwt/v2"

func GenerateClusterAccount(name string, op *Set[jwt.OperatorClaims]) (acct Set[jwt.AccountClaims], err error) {
	if acct, err = NewAccountSet(); err != nil {
		return
	}

	acct.Claims = jwt.NewAccountClaims(acct.PubKey)
	acct.Claims.Name = "Cluster - " + name
	acct.Claims.Issuer = op.PubKey

	acct.Claims.Imports = jwt.Imports{
		&jwt.Import{
			Name:    "Nix Binary Cache",
			Subject: "nits.cache",
			Type:    jwt.Service,
		},
	}

	return
}
