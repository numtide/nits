package main

import (
	"github.com/alecthomas/kong"
	"github.com/numtide/nits/internal/cmd"
)

func main() {
	ctx := kong.Parse(&cmd.Cli)
	ctx.FatalIfErrorf(ctx.Run())
}
