package nix

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"os/exec"

	"github.com/inconshreveable/log15"
	"github.com/juju/errors"
	"github.com/numtide/nits/pkg/server"
)

const (
	ErrorMalformedClosure = errors.ConstError("closure is malformed")
)

type outLogger struct {
	buf []byte
	log log15.Logger
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

func CurrentSystemClosure() (string, error) {
	return os.Readlink("/run/current-system")
}

func runCmd(name string, args []string, log log15.Logger) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = outLogger{log: log.New("output", "stdout")}
	cmd.Stderr = outLogger{log: log.New("output", "stderr")}
	return cmd.Run()
}

func CopyToBinaryCache(cacheAddr net.Addr, path string, log log15.Logger) error {
	args := []string{
		"copy",
		"-v",
		"--log-format", "raw",
		"--to", fmt.Sprintf("http://%s?compression=zstd", cacheAddr.String()),
		path,
	}
	return runCmd("nix", args, log)
}

func CopyFromBinaryCache(cacheAddr net.Addr, path string, log log15.Logger) error {
	args := []string{
		"copy",
		"-v",
		"--refresh",
		"--log-format", "raw",
		"--from", fmt.Sprintf("http://%s?compression=zstd", cacheAddr.String()),
		path,
	}
	log.Info("copying from binary cache", "args", args)
	return runCmd("nix", args, log)
}

func SwitchToConfiguration(config *server.Deployment, dryRun bool, log log15.Logger) error {
	binPath := config.Closure + "/bin/switch-to-configuration"
	_, err := os.Stat(binPath)
	if err != nil {
		return ErrorMalformedClosure
	}

	args := []string{config.Action.String()}

	if dryRun {
		log.Info("dry run", "name", binPath, "args", args)
		return nil
	}
	return runCmd(binPath, args, log)
}
