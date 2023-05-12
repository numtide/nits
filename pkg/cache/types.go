package cache

import (
	"fmt"
	"io"
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
