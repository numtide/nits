{
  inputs,
  lib,
  ...
}: {
  imports = [
    inputs.devshell.flakeModule
    inputs.process-compose-flake.flakeModule
  ];

  config.perSystem = {
    pkgs,
    config,
    ...
  }: let
    inherit (pkgs.stdenv) isLinux isDarwin;
  in {
    config.devshells.default = {
      env = [
        {
          name = "GOROOT";
          value = pkgs.go + "/share/go";
        }
        {
          name = "LD_LIBRARY_PATH";
          value = "$DEVSHELL_DIR/lib";
        }
      ];

      packages = with lib;
        mkMerge [
          [
            # golang
            pkgs.go
            pkgs.go-tools
            pkgs.delve
            pkgs.golangci-lint

            pkgs.openssl

            pkgs.qemu-utils

            pkgs.statix
          ]
          # platform dependent CGO dependencies
          (mkIf isLinux [
            pkgs.gcc
          ])
          (mkIf isDarwin [
            pkgs.darwin.cctools
          ])
        ];

      commands = let
        category = "docs";
      in [
        {
          inherit category;
          package = pkgs.vhs;
          help = "generate terminal gifs";
        }
        {
          inherit category;
          help = "regenerate gifs for docs";
          name = "gifs";
          command = ''
            set -xeuo pipefail

            for tape in $PRJ_ROOT/docs/vhs/*; do
                vhs $tape -o "$PRJ_ROOT/docs/assets/$(basename $tape .tape).gif"
            done
          '';
        }
      ];
    };
  };
}
