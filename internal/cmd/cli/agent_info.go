package cli

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/juju/errors"
	"github.com/nats-io/nats.go"
	"github.com/numtide/nits/internal/cmd"
	"github.com/numtide/nits/pkg/agent"
	"github.com/numtide/nits/pkg/agent/info"
	nutil "github.com/numtide/nits/pkg/nats"
	"github.com/shirou/gopsutil/v3/host"
)

type agentInfoCmd struct {
	Nats nutil.CliOptions `embed:"" prefix:"nats-"`
	Name string           `arg:""`

	All   bool `help:"Include all available agent info"`
	Host  bool `help:"Include information about the host machine"`
	Nix   bool `help:"Include information about the version of Nix installed on the host machine"`
	NixOS bool `name:"nixos" help:"Include information about the host machine's NixOS config'"`
	Load  bool `help:"Include load information about the host machine"`
}

func (c *agentInfoCmd) Run() error {
	Cmd.Log.ConfigureLog()

	return cmd.Run(func(ctx context.Context) (err error) {
		var (
			conn    *nats.Conn
			encoded *nats.EncodedConn
		)

		if conn, err = c.Nats.Connect(); err != nil {
			return
		} else if encoded, err = nats.NewEncodedConn(conn, nats.JSON_ENCODER); err != nil {
			return
		}

		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		var agents []*info.Response
		if agents, err = agent.List(ctx, conn); err != nil {
			return err
		}

		var nkey string
		for _, a := range agents {
			if a.Name == c.Name {
				elapsed := time.Now().Sub(a.LastSeen)
				if elapsed > 10*time.Second {
					return errors.Errorf("agent has not been seen in %v", elapsed)
				}
				nkey = a.NKey
				break
			}
		}

		if nkey == "" {
			return errors.Errorf("no agent with the name '%s' has ever reported in", c.Name)
		}

		req := info.Request{
			Host:  c.All || c.Host,
			Load:  c.All || c.Load,
			Nix:   c.All || c.Nix,
			NixOS: c.All || c.NixOS,
		}

		var resp info.Response
		if err = info.Get(encoded, nkey, req, &resp, 10*time.Second); err != nil {
			return err
		}

		printAgentSummary(&resp)
		printNix(resp.Nix)
		printNixos(resp.NixOS)
		printAgentHost(resp.Host)
		printAgentLoad(resp.Load)

		return
	})
}

func printAgentSummary(agent *info.Response) {
	println(sectionHeaderStyle.Render(fmt.Sprintf("Summary for agent %s:", agent.Name)))
	println()
	kvPrintln("Name:", agent.Name)
	kvPrintln("NKey:", agent.NKey)
	kvPrintln("Subject:", agent.Subject)
}

func printAgentHost(host *host.InfoStat) {
	if host == nil {
		return
	}

	println()
	println(sectionHeaderStyle.Render("Host:"))
	println()

	kvPrintln("Hostname:", host.Hostname)
	kvPrintln("Uptime:", (time.Duration(int64(host.Uptime)) * time.Second).String())
	kvPrintln("BootTime:", time.Unix(int64(host.BootTime), 0).Format(time.RFC1123Z))
	kvPrintln("Procs:", strconv.FormatUint(host.Procs, 10))
	kvPrintln("OS:", host.OS)
	kvPrintln("Platform:", host.Platform)
	kvPrintln("Platform Family:", host.PlatformFamily)
	kvPrintln("Platform Version:", host.PlatformVersion)
	kvPrintln("Kernel Version:", host.KernelVersion)
	kvPrintln("Kernel Arch:", host.KernelArch)
	kvPrintln("Virtualization System:", host.VirtualizationSystem)
	kvPrintln("Virtualization Role:", host.VirtualizationRole)
	kvPrintln("Host ID:", host.HostID)
}

func printNix(nix *info.Nix) {
	if nix == nil {
		return
	}

	println()
	println(sectionHeaderStyle.Render("Nix Summary:"))
	println()
	kvPrintln("System:", nix.Info.System)
	kvPrintln("Version:", nix.Info.Version)
	kvPrintln("Multi-user:", strconv.FormatBool(nix.Info.MultiUser))
	kvPrintln("Nixpkgs:", nix.Info.Nixpkgs)

	println()
	println(sectionHeaderStyle.Render("Nix Config:"))
	println()

	for k, v := range nix.Config {
		kvPrintln(k+":", v)
	}
}

func printNixos(nixos *info.NixOS) {
	if nixos == nil {
		return
	}

	println()
	println(sectionHeaderStyle.Render("NixOS Summary:"))
	println()

	kvPrintln("Version:", nixos.Version)
	kvPrintln("Current system:", nixos.CurrentSystem)
}

func printAgentLoad(load *info.Load) {
	if load == nil {
		return
	}

	println()
	println(sectionHeaderStyle.Render("Load:"))
	println()

	if load.Avg != nil {
		kvPrintln("Avg:", fmt.Sprintf("(1m) %.2f (5m) %.2f (15m) %.2f", load.Avg.Load1, load.Avg.Load5, load.Avg.Load15))
		kvPrintln("Procs Total:", strconv.Itoa(load.Misc.ProcsTotal))
		kvPrintln("Procs Created:", strconv.Itoa(load.Misc.ProcsCreated))
		kvPrintln("Procs Running:", strconv.Itoa(load.Misc.ProcsRunning))
		kvPrintln("Procs Blocked:", strconv.Itoa(load.Misc.ProcsBlocked))
		kvPrintln("Ctxt:", strconv.Itoa(load.Misc.Ctxt)) // todo what does this measure?
	}
}
