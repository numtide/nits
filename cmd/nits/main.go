package main

import (
	"github.com/alecthomas/kong"
	"github.com/brianmcgee/kilgrave/internal/cmd"
)

func main() {
	ctx := kong.Parse(&cmd.Cli)
	ctx.FatalIfErrorf(ctx.Run())
}
