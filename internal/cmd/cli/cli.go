package cli

import "github.com/numtide/nits/internal/cmd"

var Cmd struct {
	Log cmd.LogOptions `embed:""`

	Agent struct {
		Add    agentAdd    `cmd:"" help:"Add an agent to a cluster"`
		List   agentList   `cmd:"" name:"ls" help:"List agents within a cluster"`
		Info   agentInfo   `cmd:"" help:"Show info about an agent"`
		Logs   agentLogs   `cmd:"" help:"Show logs for an agent"`
		Deploy agentDeploy `cmd:"" help:"Deploy to an agent"`
	} `cmd:"" help:"Agent related functions"`

	Cluster struct {
		Add clusterAdd `cmd:""`
	} `cmd:"" help:"Cluster related functions"`
}
