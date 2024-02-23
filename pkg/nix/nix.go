package nix

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/shirou/gopsutil/v3/host"

	"github.com/nix-community/go-nix/pkg/storepath"

	"github.com/juju/errors"
)

const (
	ErrorMalformedClosure = errors.ConstError("closure is malformed")
)

var infoRegex = regexp.MustCompile(`^system: "(.*?)", multi-user\?: (.*?), version: (.*?),.*$`)

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

func IsHostNixOS() (bool, error) {
	platform, _, _, err := host.PlatformInformation()
	return "nixos" == platform, err
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

func GetSystem() (path string, err error) {
	return os.Readlink("/run/current-system")
}

func GetInfo() (info *Info, err error) {
	cmd := exec.Command("/run/current-system/sw/bin/nix-info")
	var b []byte
	if b, err = cmd.Output(); err != nil {
		return
	}

	// we need to strip a newline from the end
	matches := infoRegex.FindStringSubmatch(string(b[:len(b)-1]))
	if len(matches) != 4 {
		return nil, errors.Errorf("failed to parse nix-info output: %s", string(b))
	}

	info = &Info{
		System:    matches[1],
		MultiUser: "yes" == matches[2],
		Version:   matches[3],
	}

	return
}

func GetNixOSVersion() (string, error) {
	cmd := exec.Command("/run/current-system/sw/bin/nixos-version")
	if b, err := cmd.Output(); err != nil {
		return "", err
	} else {
		// strip a new line at the end
		return string(b[:len(b)-1]), nil
	}
}

func runCmd(name string, args []string, env []string, ctx context.Context) (err error) {
	cmd := exec.Command(name, args...)
	cmd.Env = env
	cmd.Stdout = GetStdOut(ctx)
	cmd.Stderr = GetStdErr(ctx)

	if _, err = cmd.Stderr.Write([]byte(cmd.String() + "\n")); err != nil {
		return
	} else {
		return cmd.Run()
	}
}

func Build(path *storepath.StorePath, env []string, ctx context.Context) error {
	return runCmd("nix", []string{"build", path.Absolute()}, env, ctx)
}

func SetSystem(path *storepath.StorePath, ctx context.Context) error {
	args := []string{
		"--profile", "/nix/var/nix/profiles/system",
		"--set", path.Absolute(),
	}
	return runCmd("nix-env", args, nil, ctx)
}

func Switch(closure *storepath.StorePath, action string, ctx context.Context) error {
	binPath := closure.Absolute() + "/bin/switch-to-configuration"
	_, err := os.Stat(binPath)
	if err != nil {
		return ErrorMalformedClosure
	}

	args := []string{action}

	return runCmd(binPath, args, nil, ctx)
}
