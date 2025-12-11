{
  description = "Nix & NATS";

  nixConfig = {
    extra-substituters = [
      "https://cache.numtide.com"
      "https://nix-community.cachix.org"
    ];
    extra-trusted-public-keys = [
      "niks3.numtide.com-1:DTx8wZduET09hRmMtKdQDxNNthLQETkc/yaX7M4qK0g="
      "nix-community.cachix.org-1:mB9FSh9qf2dCimDSUo8Zy7bkq5CX+/rkCWyvRCYg3Fs="
    ];
  };

  inputs = {
    srvos.url = "github:numtide/srvos";
    nixpkgs.follows = "srvos/nixpkgs";
    blueprint.url = "github:numtide/blueprint";
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
    };
    nix-lib = {
      url = "github:brianmcgee/nix-lib";
      inputs = {
        nixpkgs.follows = "srvos/nixpkgs";
        treefmt-nix.follows = "treefmt-nix";
      };
    };
    nix-filter.url = "github:numtide/nix-filter";
  };

  outputs = inputs:
    inputs.blueprint {
      inherit inputs;
      prefix = ./nix;
      systems = [
        "aarch64-linux"
        "x86_64-linux"
      ];
    };
}
