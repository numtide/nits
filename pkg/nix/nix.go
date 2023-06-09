package nix

import (
	"fmt"
	"net"
	"os"
	"os/exec"

	"github.com/juju/errors"
	"github.com/numtide/nits/pkg/guvnor"
)

const (
	ErrorMalformedClosure = errors.ConstError("closure is malformed")
)

func CurrentSystemClosure() (string, error) {
	return os.Readlink("/run/current-system")
}

func CopyToBinaryCache(cacheAddr net.Addr, path string) (stdOut []byte, stdErr []byte, err error) {
	args := []string{
		"copy",
		"-v",
		"--to",
		fmt.Sprintf("http://%s?compression=zstd", cacheAddr.String()),
		path,
	}
	cmd := exec.Command("nix", args...)

	stdOut, err = cmd.Output()
	if exitError, ok := err.(*exec.ExitError); ok {
		stdErr = exitError.Stderr
	}

	return
}

func CopyFromBinaryCache(cacheAddr net.Addr, path string) (stdOut []byte, stdErr []byte, err error) {
	args := []string{
		"copy",
		"-v",
		"--from",
		fmt.Sprintf("http://%s?compression=zstd", cacheAddr.String()),
		"--refresh",
		path,
	}
	cmd := exec.Command("nix", args...)

	stdOut, err = cmd.Output()
	if exitError, ok := err.(*exec.ExitError); ok {
		stdErr = exitError.Stderr
	}

	return
}

func SwitchToConfiguration(config *guvnor.Deployment, dryRun bool) (stdOut []byte, stdErr []byte, err error) {
	binPath := config.Closure + "/bin/switch-to-configuration"
	_, err = os.Stat(binPath)
	if err != nil {
		return nil, nil, ErrorMalformedClosure
	}

	cmd := exec.Command(binPath, config.Action.String())

	if dryRun {
		return []byte(fmt.Sprintf("dry-run: %s", cmd.String())), nil, nil
	}

	stdOut, err = cmd.Output()
	if exitError, ok := err.(*exec.ExitError); ok {
		stdErr = exitError.Stderr
	}

	return
}
