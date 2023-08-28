package nats

import (
	"io"
	"os"

	"github.com/nats-io/jwt/v2"
)

type claimsDecoder[T any] func(string) (*T, error)

func ParseClaims[T any](str string, decoder claimsDecoder[T]) (claims *T, err error) {
	return decoder(str)
}

func ParseUserClaims(str string) (*jwt.UserClaims, error) {
	return ParseClaims[jwt.UserClaims](str, jwt.DecodeUserClaims)
}

func ParseAccountClaims(str string) (*jwt.AccountClaims, error) {
	return ParseClaims[jwt.AccountClaims](str, jwt.DecodeAccountClaims)
}

func ParseOperatorClaims(str string) (*jwt.OperatorClaims, error) {
	return ParseClaims[jwt.OperatorClaims](str, jwt.DecodeOperatorClaims)
}

func ReadClaims[T any](path string, decoder claimsDecoder[T]) (claims *T, err error) {
	var b []byte
	if b, err = os.ReadFile(path); err != nil {
		return
	}
	return ParseClaims[T](string(b), decoder)
}

func ReadClaimsFile[T any](file *os.File, decoder claimsDecoder[T]) (claims *T, err error) {
	var b []byte
	if b, err = io.ReadAll(file); err != nil {
		return
	}
	return ParseClaims[T](string(b), decoder)
}

func ReadUserClaims(path string) (*jwt.UserClaims, error) {
	return ReadClaims[jwt.UserClaims](path, jwt.DecodeUserClaims)
}

func ReadUserClaimsFile(file *os.File) (*jwt.UserClaims, error) {
	return ReadClaimsFile[jwt.UserClaims](file, jwt.DecodeUserClaims)
}

func ReadAccountClaims(path string) (*jwt.AccountClaims, error) {
	return ReadClaims[jwt.AccountClaims](path, jwt.DecodeAccountClaims)
}

func ReadOperatorClaims(path string) (*jwt.OperatorClaims, error) {
	return ReadClaims[jwt.OperatorClaims](path, jwt.DecodeOperatorClaims)
}
