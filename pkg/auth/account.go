package auth

import (
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
)

func ReadAccountJwt(path string) (s *Set[jwt.AccountClaims], err error) {
	var token string
	if token, err = ReadJwt(path); err != nil {
		return
	}

	var claims *jwt.AccountClaims
	if claims, err = jwt.DecodeAccountClaims(token); err != nil {
		return
	}

	s = &Set[jwt.AccountClaims]{
		PubKey: claims.Subject,
		Jwt:    token,
		Claims: claims,
	}

	return
}

func ReadAccountCredentials(path string) (s *Set[jwt.AccountClaims], err error) {
	s = &Set[jwt.AccountClaims]{}
	if s.KP, s.Jwt, err = ReadCredentials(path); err != nil {
		return
	} else if s.PubKey, err = s.KP.PublicKey(); err != nil {
		return
	} else if s.Claims, err = jwt.DecodeAccountClaims(s.Jwt); err != nil {
		return
	}
	return
}

func NewAccountSet() (Set[jwt.AccountClaims], error) {
	return newSet[jwt.AccountClaims](nkeys.CreateAccount)
}

func NewClusterAccount(name string, dataDir string) (acct Set[jwt.AccountClaims], err error) {
	var op *Set[jwt.OperatorClaims]
	if op, err = ReadOperatorCredentials(dataDir + "/creds/operator-signer.creds"); err != nil {
		return
	}

	if acct, err = NewAccountSet(); err != nil {
		return acct, err
	}
	acct.Claims = jwt.NewAccountClaims(acct.PubKey)
	acct.Claims.Name = name
	acct.Claims.Issuer = op.PubKey

	if err = acct.EncodeClaims(op.KP); err != nil {
		return
	}

	if err = acct.WriteJwt(dataDir+"/jwt", name+".jwt"); err != nil {
		return
	}

	if err = acct.WriteCredentials(dataDir+"/creds", name+".creds"); err != nil {
		return
	}

	return
}
