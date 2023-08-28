package nats

import (
	"io"
	"os"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
)

func ReadCredentials[T any](path string, decoder claimsDecoder[T]) (nkey nkeys.KeyPair, claims *T, err error) {
	var file *os.File
	if file, err = os.Open(path); err != nil {
		return
	}

	return ReadCredentialsFile(file, decoder)
}

func ReadCredentialsFile[T any](file *os.File, decoder claimsDecoder[T]) (nkey nkeys.KeyPair, claims *T, err error) {
	var b []byte
	var str string

	if b, err = io.ReadAll(file); err != nil {
		return
	}
	if nkey, err = nkeys.ParseDecoratedNKey(b); err != nil {
		return
	} else if str, err = nkeys.ParseDecoratedJWT(b); err != nil {
		return
	}

	claims, err = ParseClaims(str, decoder)
	return
}

func ReadUserCredentials(path string) (nkey nkeys.KeyPair, claims *jwt.UserClaims, err error) {
	return ReadCredentials[jwt.UserClaims](path, jwt.DecodeUserClaims)
}

func ReadUserCredentialsFile(file *os.File) (nkey nkeys.KeyPair, claims *jwt.UserClaims, err error) {
	return ReadCredentialsFile[jwt.UserClaims](file, jwt.DecodeUserClaims)
}
