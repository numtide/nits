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

            # nats
            pkgs.nsc
            pkgs.nkeys
            pkgs.natscli
            pkgs.nats-top

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
    };
  };
}
