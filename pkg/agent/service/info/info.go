package info

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go/micro"
	"github.com/numtide/nits/pkg/agent/util"
	"github.com/numtide/nits/pkg/subject"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
)

var agentSubject string

func Init(ctx context.Context) (err error) {
	conn := util.GetConn(ctx)
	nkey := util.GetNKey(ctx)

	agentSubject = subject.AgentWithNKey(nkey)

	_, err = micro.AddService(conn, micro.Config{
		Name:        "AgentInfo",
		Version:     "0.0.1",
		Description: "Information about an agent and the machine it is running on",
		Endpoint: &micro.EndpointConfig{
			Subject: subject.AgentService(nkey, "INFO"),
			Handler: micro.HandlerFunc(handler),
		},
	})

	return
}

func handler(req micro.Request) {
	var err error
	var request Request

	if err = json.Unmarshal(req.Data(), &request); err != nil {
		// todo handle error
	}

	resp := Response{
		Subject: agentSubject,
	}

	// todo handle errors
	if request.All || request.Cpus {
		resp.Cpus, _ = cpu.Info()
	}

	if request.All || request.Host {
		resp.Host, _ = host.Info()
	}

	if request.All || request.Load {
		resp.Load = &Load{}
		resp.Load.Avg, _ = load.Avg()
		resp.Load.Misc, _ = load.Misc()
	}

	if request.All || request.Disk {
		resp.Disk = &Disk{}
		resp.Disk.Partitions, _ = disk.Partitions(true)
	}

	if request.All || request.Memory {
		resp.Memory = &Memory{}
		resp.Memory.Swap, _ = mem.SwapMemory()
		resp.Memory.SwapDevices, _ = mem.SwapDevices()
		resp.Memory.Virtual, _ = mem.VirtualMemory()
	}

	var data []byte
	if data, err = json.Marshal(resp); err != nil {
		// todo
	}

	_ = req.Respond(data)
}
