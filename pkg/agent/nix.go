package agent

import (
	"fmt"
	"net"
	"os/exec"
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
