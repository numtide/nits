package nix

type Info struct {
	System    string `json:"system"`
	MultiUser bool   `json:"multi-user"`
	Version   string `json:"version"`
	Nixpkgs   string `json:"nixpkgs"`
	// todo add channels
}
