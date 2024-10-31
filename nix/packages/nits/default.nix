{
  flake,
  pkgs,
  inputs,
  perSystem,
  pname,
  ...
}: let
  inherit (pkgs) lib;
  filter = inputs.nix-filter.lib;
in
  perSystem.gomod2nix.buildGoApplication rec {
    inherit pname;
    version = "0.0.1+dev";

    runtimeInputs = with pkgs; [nsc natscli];

    src = filter {
      root = flake;
      include = [
        "go.mod"
        "go.sum"
        "pkg"
        "internal"
        "cmd"
      ];
    };

    modules = ./gomod2nix.toml;

    ldflags = [
      "-X 'build.Name=${pname}'"
      "-X 'build.Version=${version}'"
    ];

    meta = with lib; {
      description = "Nix & NATS";
      homepage = "https://github.com/numtide/nits";
      license = licenses.mit;
      mainProgram = "nits";
    };
  }
