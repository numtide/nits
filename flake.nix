{
  description = "Nix & NATS";

  nixConfig = {
    extra-substituters = [
      "https://numtide.cachix.org"
      "https://nix-community.cachix.org"
    ];
    extra-trusted-public-keys = [
      "numtide.cachix.org-1:2ps1kLBUWjxIneOy1Ik6cQjb41X0iXVXeHigGmycPPE="
      "nix-community.cachix.org-1:mB9FSh9qf2dCimDSUo8Zy7bkq5CX+/rkCWyvRCYg3Fs="
    ];
  };

  inputs = {
    srvos.url = "github:numtide/srvos";
    nixpkgs.follows = "srvos/nixpkgs";
    flake-parts.follows = "srvos/flake-parts";
    flake-root.follows = "nix-lib/flake-root";
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "srvos/nixpkgs";
    };
    devshell = {
      url = "github:numtide/devshell";
      inputs.nixpkgs.follows = "srvos/nixpkgs";
    };
    process-compose-flake.url = "github:Platonic-Systems/process-compose-flake";
    harmonia = {
      url = "github:nix-community/harmonia";
      inputs = {
        nixpkgs.follows = "srvos/nixpkgs";
        flake-parts.follows = "flake-parts";
        treefmt-nix.follows = "treefmt-nix";
      };
    };
    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs = {
        nixpkgs.follows = "srvos/nixpkgs";
        flake-utils.follows = "devshell/flake-utils";
      };
    };
    flake-linter = {
      url = "github:mic92/flake-linter";
      inputs.flake-parts.follows = "flake-parts";
    };
    nix-lib = {
      url = "github:brianmcgee/nix-lib";
      inputs = {
        nixpkgs.follows = "srvos/nixpkgs";
        flake-parts.follows = "flake-parts";
        treefmt-nix.follows = "treefmt-nix";
      };
    };
    nix-filter.url = "github:numtide/nix-filter";
  };

  outputs = inputs @ {
    flake-parts,
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
      imports = [./nix];
      systems = [
        "x86_64-linux"
        "aarch64-linux"
      ];
    };
}
