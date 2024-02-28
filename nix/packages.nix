{
  self,
  inputs,
  ...
}: {
  perSystem = {
    lib,
    pkgs,
    ...
  }: let
    filter = inputs.nix-filter.lib;
  in {
    packages = rec {
      nits = pkgs.buildGoApplication rec {
        pname = "nits";
        version = "0.0.1+dev";

        runtimeInputs = with pkgs; [nsc natscli];

        src = filter {
          root = ../.;
          include = [
            "go.mod"
            "go.sum"
            "pkg"
            "internal"
            "cmd"
          ];
        };

        modules = ../gomod2nix.toml;

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
      };

      default = nits;
    };
  };

  flake.overlays.default = final: _prev: {
    inherit (self.packages.${final.system}) nits;
  };
}
