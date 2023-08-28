package cli

var Cmd struct {
	Add struct {
		Cache   addCacheCmd   `cmd:""`
		Cluster addClusterCmd `cmd:""`
		Agent   addAgentCmd   `cmd:""`
	} `cmd:"" help:"Add assets such as clusters and agents"`

	Deploy deployCmd `cmd:"" help:"deploy a NixOS closure to one or more agents"`
}
