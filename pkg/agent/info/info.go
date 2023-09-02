package info

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/juju/errors"
	"github.com/nats-io/nats.go"

	"github.com/nats-io/jwt/v2"

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
	NKey   string
	Claims *jwt.UserClaims
	logger *log.Logger
)

func Init(ctx context.Context) (err error) {
	NKey = util.GetNKey(ctx)
	Claims = util.GetClaims(ctx)
	conn := util.GetConn(ctx)

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

	// send a basic info package every second to the registry subject

	info := Response{NKey: NKey, Name: Claims.Name, Subject: subject.AgentWithNKey(NKey)}

	var heartbeat []byte
	if heartbeat, err = json.Marshal(info); err != nil {
		return err
	}

	go func() {
		ticker := time.Tick(1 * time.Second)
		subj := subject.AgentRegistration(NKey)

		var (
			err error
			msg *nats.Msg
		)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker:
				msg = nats.NewMsg(subj)
				msg.Data = heartbeat
				// conflate updates
				msg.Header.Set(nats.MsgRollup, nats.MsgRollupSubject)

				if err = conn.PublishMsg(msg); err != nil {
					log.Error("failed to publish registry heartbeat", "error", err)
				}
			}
		}
	}()

	return
}

func handler(req micro.Request) {
	var (
		err      error
		request  Request
		response *Response
	)

	if len(req.Data()) > 0 {
		// we accept empty request data as a default request
		if err = json.Unmarshal(req.Data(), &request); err != nil {
			_ = req.Error("500", fmt.Sprintf("Failed to unmarshal request: %s", err), req.Data())
			return
		}
	}

	if response, err = info(&request); err != nil {
		_ = req.Error("500", err.Error(), nil)
		return
	}

	var data []byte
	if data, err = json.Marshal(response); err != nil {
		_ = req.Error("500", fmt.Sprintf("Failed to marshal response: %s", err), nil)
		return
	}

	if err = req.Respond(data); err != nil {
		logger.Error("failed to respond", "error", err)
	}
}

func info(req *Request) (resp *Response, err error) {
	resp = &Response{
		NKey:    NKey,
		Name:    Claims.Name,
		Subject: subject.AgentWithNKey(NKey),
	}

	if req.All || req.Cpus {
		if resp.Cpus, err = cpu.Info(); err != nil {
			return nil, errors.Annotate(err, "failed to retrieve cpu info")
		}
	}

	if req.All || req.Host {
		if resp.Host, err = host.Info(); err != nil {
			return nil, errors.Annotate(err, "failed to retrieve host info")
		}
	}

	if req.All || req.Load {
		resp.Load = &Load{}
		if resp.Load.Avg, err = load.Avg(); err != nil {
			return nil, errors.Annotate(err, "failed to retrieve load avg")
		}

		if resp.Load.Misc, err = load.Misc(); err != nil {
			return nil, errors.Annotate(err, "failed to retrieve miscellaneous load info")
		}
	}

	if req.All || req.Disk {
		resp.Disk = &Disk{}
		if resp.Disk.Partitions, err = disk.Partitions(true); err != nil {
			return nil, errors.Annotate(err, "failed to retrieve disk partitions")
		}
	}

	if req.All || req.Memory {
		resp.Memory = &Memory{}
		if resp.Memory.Swap, err = mem.SwapMemory(); err != nil {
			return nil, errors.Annotate(err, "failed to retrieve swap info")
		}
		if resp.Memory.SwapDevices, err = mem.SwapDevices(); err != nil {
			return nil, errors.Annotate(err, "Failed to retrieve swap devices")
		}
		if resp.Memory.Virtual, err = mem.VirtualMemory(); err != nil {
			return nil, errors.Annotate(err, "failed to retrieve virtual memory")
		}
	}

	return
}

func GetWithContext(ctx context.Context, conn *nats.EncodedConn, nkey string, req Request) (resp *Response, err error) {
	err = conn.RequestWithContext(ctx, subject.AgentService(nkey, "INFO"), req, &resp)
	return
}
