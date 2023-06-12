package main

import (
	"github.com/alecthomas/kong"
	"github.com/numtide/nits/internal/cmd/agent"
)

func main() {
	ctx := kong.Parse(&agent.Cmd)
	ctx.FatalIfErrorf(ctx.Run())
}
