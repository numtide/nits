package main

import (
	"github.com/alecthomas/kong"
	"github.com/numtide/nits/internal/cmd/guvnor"
)

func main() {
	ctx := kong.Parse(&guvnor.Cmd)
	ctx.FatalIfErrorf(ctx.Run())
}
