package cli

var Cmd struct {
	List struct {
		Agents listAgentsCmd `cmd:""`
	} `cmd:"" help:"List assets such as agents"`

	Add struct {
		Cache   addCacheCmd   `cmd:""`
		Cluster addClusterCmd `cmd:""`
		Agent   addAgentCmd   `cmd:""`
	} `cmd:"" help:"Add assets such as clusters and agents"`

	Agent struct {
		Info agentInfoCmd `cmd:""`
	} `cmd:"" help:"Agent related functions"`

	Deploy deployCmd `cmd:"" help:"deploy a NixOS closure to one or more agents"`
}
