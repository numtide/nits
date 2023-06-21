package main

import (
	"github.com/alecthomas/kong"
	"github.com/numtide/nits/internal/cmd/server"
)

func main() {
	ctx := kong.Parse(&server.Cmd)
	ctx.FatalIfErrorf(ctx.Run())
}
