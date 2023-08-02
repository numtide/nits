package auth

import (
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
)

func ReadOperatorJwt(path string) (result *Set[jwt.OperatorClaims], err error) {
	var token string
	if token, err = ReadJwt(path); err != nil {
		return
	}

	var claims *jwt.OperatorClaims
	if claims, err = jwt.DecodeOperatorClaims(token); err != nil {
		return
	}

	result = &Set[jwt.OperatorClaims]{
		PubKey: claims.Subject,
		Jwt:    token,
		Claims: claims,
	}

	return
}

func ReadOperatorCredentials(path string) (s *Set[jwt.OperatorClaims], err error) {
	s = &Set[jwt.OperatorClaims]{}
	if s.KP, s.Jwt, err = ReadCredentials(path); err != nil {
		return
	} else if s.PubKey, err = s.KP.PublicKey(); err != nil {
		return
	} else if s.Claims, err = jwt.DecodeOperatorClaims(s.Jwt); err != nil {
		return
	}
	return
}

func NewOperatorSet() (s Set[jwt.OperatorClaims], err error) {
	return newSet[jwt.OperatorClaims](nkeys.CreateOperator)
}
