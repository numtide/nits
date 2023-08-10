package cli

import (
	"embed"
	"io"
	"io/fs"
	"os"

	"github.com/ztrue/shutdown"
)

//go:embed streams
var streamConfig embed.FS

func openResourceLocally(filesystem embed.FS, filename string) (f *os.File, err error) {
	var resource fs.File
	if resource, err = filesystem.Open(filename); err != nil {
		return
	}

	if f, err = os.CreateTemp("", "nits-resource-"); err != nil {
		return
	}

	if _, err = io.Copy(f, resource); err != nil {
		return
	} else if err = f.Close(); err != nil {
		return
	}

	// delete on process exit
	shutdown.Add(func() {
		_ = os.Remove(f.Name())
	})

	return
}
