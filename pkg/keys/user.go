package keys

import (
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
)

func ReadUserJwt(path string) (result *Set[jwt.UserClaims], err error) {
	var token string
	if token, err = ReadJwt(path); err != nil {
		return
	}

	var claims *jwt.UserClaims
	if claims, err = jwt.DecodeUserClaims(token); err != nil {
		return
	}

	result = &Set[jwt.UserClaims]{
		PubKey: claims.Subject,
		Jwt:    token,
		Claims: claims,
	}

	return
}

func ReadUserCredentials(path string) (s *Set[jwt.UserClaims], err error) {
	if s.KP, s.Jwt, err = ReadCredentials(path); err != nil {
		return
	} else if s.PubKey, err = s.KP.PublicKey(); err != nil {
		return
	} else if s.Claims, err = jwt.DecodeUserClaims(s.Jwt); err != nil {
		return
	}
	return
}

func NewUserSet() (s Set[jwt.UserClaims], err error) {
	return newSet[jwt.UserClaims](nkeys.CreateUser)
}
