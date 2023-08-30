package nix

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/nix-community/go-nix/pkg/nixpath"
	"github.com/numtide/nits/pkg/types"

	"github.com/charmbracelet/log"

	"github.com/juju/errors"
)

const (
	ErrorMalformedClosure = errors.ConstError("closure is malformed")
)

func Config() (config map[string]string, err error) {
	cmd := exec.Command("nix", "show-config")

	config = make(map[string]string)

	var b []byte
	if b, err = cmd.Output(); err != nil {
		return
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(b))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		components := strings.SplitN(line, " = ", 2)
		if len(components) != 2 {
			return nil, errors.Errorf("malformed line in output: %s", line)
		}
		config[components[0]] = components[1]
	}

	return
}

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

func CurrentSystem() (*nixpath.NixPath, error) {
	path, err := os.Readlink("/run/current-system")
	if err != nil {
		return nil, err
	}
	return nixpath.FromString(path)
}

func runCmd(name string, args []string, env []string, ctx context.Context) error {
	logger := log.FromContext(ctx)

	logger.Info(name, "args", strings.Join(args, " "), "env", env)

	prefix := logger.GetPrefix()
	logger.SetPrefix(name)

	// reset prefix afterwards
	defer logger.SetPrefix(prefix)

	cmd := exec.Command(name, args...)
	cmd.Env = env
	cmd.Stdout = outLogger{log: logger.With("output", "stdout")}
	cmd.Stderr = outLogger{log: logger.With("output", "stderr")}

	// todo be able to interrupt a command?
	return cmd.Run()
}

func BuildSystemClosure(path *nixpath.NixPath, env []string, ctx context.Context) error {
	args := []string{
		"build", path.String(),
	}
	return runCmd("nix", args, env, ctx)
}

func SetSystemProfile(path *nixpath.NixPath, ctx context.Context) error {
	args := []string{
		"--profile", "/nix/var/nix/profiles/system",
		"--set", path.String(),
	}
	return runCmd("nix-env", args, nil, ctx)
}

func SwitchToConfiguration(config *types.Deployment, ctx context.Context) error {
	binPath := config.Closure + "/bin/switch-to-configuration"
	_, err := os.Stat(binPath)
	if err != nil {
		return ErrorMalformedClosure
	}

	args := []string{config.Action.String()}

	return runCmd(binPath, args, nil, ctx)
}
