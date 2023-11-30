{
  description = "Nix & NATS";

  nixConfig = {
    extra-substituters = [
      "https://cache.garnix.io"
      "https://nix-community.cachix.org"
    ];
    extra-trusted-public-keys = [
      "cache.garnix.io:CTFPyKSLcx5RMJKfLo5EEPUObbA78b0YQ2DTCJXqr9g="
      "nix-community.cachix.org-1:mB9FSh9qf2dCimDSUo8Zy7bkq5CX+/rkCWyvRCYg3Fs="
    ];
  };

  inputs = {
    srvos.url = "github:numtide/srvos";
    # Use the version of nixpkgs that has been tested to work with SrvOS
    nixpkgs.follows = "srvos/nixpkgs";
    flake-parts.url = "github:hercules-ci/flake-parts";
    flake-root.url = "github:srid/flake-root";
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    devshell = {
      url = "github:numtide/devshell";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    process-compose-flake.url = "github:Platonic-Systems/process-compose-flake";
    nix-serve.url = "github:edolstra/nix-serve";
    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = inputs @ {
    flake-parts,
    gomod2nix,
    nixpkgs,
    ...
  }: let
    lib = nixpkgs.lib.extend (import ./nix/lib.nix);
  in
    flake-parts.lib.mkFlake
    {
      inherit inputs;
      specialArgs = {
        inherit lib; # make custom lib available to top level functions
      };
    } {
      imports = [
        {
          perSystem = {system, ...}: {
            # customize nixpkgs instance
            _module.args.pkgs = import nixpkgs {
              inherit system;
              overlays = [
                gomod2nix.overlays.default
              ];
            };
          };
        }
        ./nix
      ];
      systems = [
        "x86_64-linux"
        "aarch64-linux"
      ];
    };
}
