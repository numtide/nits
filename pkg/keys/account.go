package keys

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
	if s.KP, s.Jwt, err = ReadCredentials(path); err != nil {
		return
	} else if s.PubKey, err = s.KP.PublicKey(); err != nil {
		return
	} else if s.Claims, err = jwt.DecodeAccountClaims(s.Jwt); err != nil {
		return
	}
	return
}

func NewAccountSet() (s Set[jwt.AccountClaims], err error) {
	return newSet[jwt.AccountClaims](nkeys.CreateAccount)
}
