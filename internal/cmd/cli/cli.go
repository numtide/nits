package cli

import "github.com/numtide/nits/internal/cmd"

var Cmd struct {
	Log cmd.LogOptions `embed:"" prefix:"log-"`

	List struct {
		Agents listAgentsCmd `cmd:""`
	} `cmd:"" help:"List assets such as agents"`

	Add struct {
		Cache   addCacheCmd   `cmd:""`
		Cluster addClusterCmd `cmd:""`
		Agent   addAgentCmd   `cmd:""`
	} `cmd:"" help:"Add assets such as clusters and agents"`

	Agent struct {
		Info   agentInfoCmd   `cmd:""`
		Logs   agentLogsCmd   `cmd:""`
		Deploy agentDeployCmd `cmd:""`
	} `cmd:"" help:"Agent related functions"`
}
