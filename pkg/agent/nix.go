package agent

import (
	"fmt"
	"github.com/juju/errors"
	"github.com/numtide/nits/pkg/guvnor"
	"net"
	"os"
	"os/exec"

	log "github.com/inconshreveable/log15"
)

const (
	ErrorMalformedClosure = errors.ConstError("closure is malformed")
)

func copyFromBinaryCache(cacheAddr net.Addr, path string) error {
	args := []string{
		"copy",
		"--from",
		fmt.Sprintf("http://%s?compression=zstd", cacheAddr.String()),
		path,
	}
	cmd := exec.Command("nix", args...)
	return cmd.Run()
}

func switchToConfiguration(config *guvnor.Deployment, dryRun bool, logger log.Logger) error {

	binPath := config.Closure + "/bin/switch-to-configuration"
	_, err := os.Stat(binPath)
	if err != nil {
		return ErrorMalformedClosure
	}

	cmd := exec.Command(binPath, config.Action.String())

	if dryRun {
		logger.Info("switching nixos configuration", "dryRun", dryRun, "args", cmd.Args)
		return nil
	}

	return cmd.Run()
}
