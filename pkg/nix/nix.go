package nix

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/nix-community/go-nix/pkg/nixpath"

	"github.com/juju/errors"
)

const (
	ErrorMalformedClosure = errors.ConstError("closure is malformed")
)

func SetStdError(ctx context.Context, writer io.Writer) context.Context {
	return context.WithValue(ctx, "stderr", writer)
}

func GetStdErr(ctx context.Context) io.Writer {
	return ctx.Value("stderr").(io.Writer)
}

func SetStdOut(ctx context.Context, writer io.Writer) context.Context {
	return context.WithValue(ctx, "stdout", writer)
}

func GetStdOut(ctx context.Context) io.Writer {
	return ctx.Value("stdout").(io.Writer)
}

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

func CurrentSystem() (*nixpath.NixPath, error) {
	path, err := os.Readlink("/run/current-system")
	if err != nil {
		return nil, err
	}
	return nixpath.FromString(path)
}

func runCmd(name string, args []string, env []string, ctx context.Context) error {
	cmd := exec.Command(name, args...)
	cmd.Env = env
	cmd.Stdout = GetStdOut(ctx)
	cmd.Stderr = GetStdErr(ctx)
	// todo be able to interrupt a command?
	return cmd.Run()
}

func Build(path *nixpath.NixPath, env []string, ctx context.Context) error {
	return runCmd("nix", []string{"build", path.String()}, env, ctx)
}

func SetSystem(path *nixpath.NixPath, ctx context.Context) error {
	args := []string{
		"--profile", "/nix/var/nix/profiles/system",
		"--set", path.String(),
	}
	return runCmd("nix-env", args, nil, ctx)
}

func Switch(closure *nixpath.NixPath, action string, ctx context.Context) error {
	binPath := closure.String() + "/bin/switch-to-configuration"
	_, err := os.Stat(binPath)
	if err != nil {
		return ErrorMalformedClosure
	}

	args := []string{action}

	return runCmd(binPath, args, nil, ctx)
}
