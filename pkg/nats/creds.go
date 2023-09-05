package nats

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"

	"github.com/nats-io/nkeys"
	"github.com/nats-io/nsc/v2/cmd"
)

func ReadCredentials(path string) (nkey nkeys.KeyPair, jwt string, err error) {
	var file *os.File
	if file, err = os.Open(path); err != nil {
		return
	}

	return ReadCredentialsFile(file)
}

func ReadCredentialsFile(file *os.File) (nkey nkeys.KeyPair, jwt string, err error) {
	var b []byte

	if b, err = io.ReadAll(file); err != nil {
		return
	}
	if nkey, err = nkeys.ParseDecoratedNKey(b); err != nil {
		return
	}

	jwt, err = nkeys.ParseDecoratedJWT(b)
	return
}

func ReadProfile(url string) (profile cmd.Profile, nkey nkeys.KeyPair, jwt string, err error) {
	var b []byte
	if b, err = exec.Command("nsc", "generate", "profile", url).Output(); err != nil {
		return
	}

	if err = json.Unmarshal(b, &profile); err != nil {
		return
	}

	nkey, jwt, err = ReadCredentials(profile.UserCreds)
	return
}
