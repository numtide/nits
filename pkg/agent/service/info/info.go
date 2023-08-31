package info

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/log"

	"github.com/nats-io/nats.go/micro"
	"github.com/numtide/nits/pkg/agent/util"
	"github.com/numtide/nits/pkg/subject"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
)

var (
	NKey string

	logger *log.Logger
)

func Init(ctx context.Context) (err error) {
	conn := util.GetConn(ctx)
	NKey = util.GetNKey(ctx)

	logger = log.Default().With("service", "info")

	_, err = micro.AddService(conn, micro.Config{
		Name:        "AgentInfo",
		Version:     "0.0.1",
		Description: "Information about an agent and the machine it is running on",
		Endpoint: &micro.EndpointConfig{
			Subject: subject.AgentService(NKey, "INFO"),
			Handler: micro.HandlerFunc(handler),
		},
	})

	return
}

func handler(req micro.Request) {
	var err error
	var request Request

	if len(req.Data()) > 0 {
		// we accept empty request data as a default request
		if err = json.Unmarshal(req.Data(), &request); err != nil {
			_ = req.Error("500", fmt.Sprintf("Failed to unmarshal request: %s", err), req.Data())
			return
		}
	}

	resp := Response{
		NKey:    NKey,
		Subject: subject.AgentWithNKey(NKey),
	}

	if request.All || request.Cpus {
		if resp.Cpus, err = cpu.Info(); err != nil {
			_ = req.Error("500", fmt.Sprintf("Failed to retrieve cpu info: %s", err), nil)
			return
		}
	}

	if request.All || request.Host {
		if resp.Host, err = host.Info(); err != nil {
			_ = req.Error("500", fmt.Sprintf("Failed to retrieve host info: %s", err), nil)
			return
		}
	}

	if request.All || request.Load {
		resp.Load = &Load{}
		if resp.Load.Avg, err = load.Avg(); err != nil {
			_ = req.Error("500", fmt.Sprintf("Failed to retrieve load avg: %s", err), nil)
			return
		}

		if resp.Load.Misc, err = load.Misc(); err != nil {
			_ = req.Error("500", fmt.Sprintf("Failed to retrieve miscellaneous load info: %s", err), nil)
			return
		}
	}

	if request.All || request.Disk {
		resp.Disk = &Disk{}
		if resp.Disk.Partitions, err = disk.Partitions(true); err != nil {
			_ = req.Error("500", fmt.Sprintf("Failed to retrieve disk partitions: %s", err), nil)
			return
		}
	}

	if request.All || request.Memory {
		resp.Memory = &Memory{}
		if resp.Memory.Swap, err = mem.SwapMemory(); err != nil {
			_ = req.Error("500", fmt.Sprintf("Failed to retrieve swap info: %s", err), nil)
			return
		}
		if resp.Memory.SwapDevices, err = mem.SwapDevices(); err != nil {
			_ = req.Error("500", fmt.Sprintf("Failed to retrieve swap devices: %s", err), nil)
			return
		}
		if resp.Memory.Virtual, err = mem.VirtualMemory(); err != nil {
			_ = req.Error("500", fmt.Sprintf("Failed to retrieve virtual memory: %s", err), nil)
			return
		}
	}

	var data []byte
	if data, err = json.Marshal(resp); err != nil {
		_ = req.Error("500", fmt.Sprintf("Failed to marshal response: %s", err), nil)
		return
	}

	if err = req.Respond(data); err != nil {
		logger.Error("failed to respond", "error", err)
	}
}
