package nix

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"

	"github.com/nix-community/go-nix/pkg/nixpath"
	"github.com/numtide/nits/pkg/types"

	"github.com/charmbracelet/log"

	"github.com/juju/errors"
)

const (
	ErrorMalformedClosure = errors.ConstError("closure is malformed")
)

type outLogger struct {
	buf []byte
	log *log.Logger
}

func (o outLogger) Write(b []byte) (n int, err error) {
	buf := o.buf

	buf = append(buf, b...)
	scanner := bufio.NewScanner(bytes.NewBuffer(buf))

	count := 0

	for scanner.Scan() {
		msg := scanner.Text()
		count += len(msg)
		o.log.Info(msg)
	}

	// resize based on number of bytes read
	o.buf = buf[count:]

	return len(b), nil
}

func CurrentSystemClosure() (*nixpath.NixPath, error) {
	path, err := os.Readlink("/run/current-system")
	if err != nil {
		return nil, err
	}
	return nixpath.FromString(path)
}

func runCmd(name string, args []string, ctx context.Context) error {
	logger := log.FromContext(ctx)
	cmd := exec.Command(name, args...)
	cmd.Stdout = outLogger{log: logger.With("output", "stdout")}
	cmd.Stderr = outLogger{log: logger.With("output", "stderr")}

	// todo be able to interrupt a command?
	return cmd.Run()
}

func SetSystemProfile(path *nixpath.NixPath, ctx context.Context) error {
	args := []string{
		"--profile", "/nix/var/nix/profiles/system",
		"--set", path.String(),
	}
	return runCmd("nix-env", args, ctx)
}

func CopyToBinaryCache(cacheAddr net.Addr, path string, ctx context.Context) error {
	args := []string{
		"copy",
		"-v",
		"--log-format", "raw",
		"--to", fmt.Sprintf("http://%s?compression=zstd", cacheAddr.String()),
		path,
	}
	return runCmd("nix", args, ctx)
}

func CopyFromBinaryCache(cacheAddr net.Addr, path string, ctx context.Context) error {
	args := []string{
		"copy",
		"--refresh",
		"--log-format", "raw",
		"--from", fmt.Sprintf("http://%s?compression=zstd", cacheAddr.String()),
		path,
	}
	log.FromContext(ctx).Info("copying from binary cache", "args", args)
	return runCmd("nix", args, ctx)
}

func SwitchToConfiguration(config *types.Deployment, ctx context.Context) error {
	binPath := config.Closure + "/bin/switch-to-configuration"
	_, err := os.Stat(binPath)
	if err != nil {
		return ErrorMalformedClosure
	}

	args := []string{config.Action.String()}

	return runCmd(binPath, args, ctx)
}
