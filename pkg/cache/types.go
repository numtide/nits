package cache

import (
	"fmt"
	"io"

	"github.com/juju/errors"
)

type Info struct {
	StoreDir      string
	WantMassQuery bool
	Priority      int
}

func (i Info) Write(w io.Writer) (err error) {
	massQuery := 0
	if i.WantMassQuery {
		massQuery = 1
	}
	str := fmt.Sprintf(
		"StoreDir: %s\nWantMassQuery: %d\nPriority %d",
		i.StoreDir,
		massQuery,
		i.Priority,
	)
	_, err = io.WriteString(w, str)
	return err
}

func compressionExtension(compression string) (result string, err error) {
	switch compression {
	case "xz":
		result = "xz"
	case "bzip2":
		result = "bz2"
	case "zstd":
		result = "zst"
	case "lzip":
		result = "lzip"
	case "lz4":
		result = "lz4"
	case "br":
		result = "br"
	default:
		err = errors.New("unexpected compression")
	}
	return
}
