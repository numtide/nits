package nixos

import (
	"github.com/nats-io/nats.go/micro"
	"github.com/numtide/nits/pkg/nix"
)

type InfoResponse struct {
	System string            `json:"system"`
	Config map[string]string `json:"config"`
}

func onInfo(req micro.Request) {
	var (
		err    error
		system string
		config map[string]string
	)

	if system, err = nix.GetSystem(); err != nil {
		_ = req.Error("500", err.Error(), nil)
		return
	}

	if config, err = nix.Config(); err != nil {
		_ = req.Error("500", err.Error(), nil)
		return
	}

	response := InfoResponse{
		System: system,
		Config: config,
	}

	if err = req.RespondJSON(response); err != nil {
		logger.Error("failed to respond", "error", err)
	}
}
