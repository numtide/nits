package main

import (
	"github.com/alecthomas/kong"
	"github.com/numtide/nits/internal/cmd/cli"
)

func main() {
	ctx := kong.Parse(&cli.Cmd)
	ctx.FatalIfErrorf(ctx.Run())
}
