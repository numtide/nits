{
  perSystem,
  pkgs,
  ...
}: let
  inherit (pkgs.stdenv) isLinux isDarwin;
  inherit (pkgs) lib;
in
  perSystem.devshell.mkShell {
    env = [
      {
        name = "GOROOT";
        value = pkgs.go + "/share/go";
      }
      {
        name = "LD_LIBRARY_PATH";
        value = "$DEVSHELL_DIR/lib";
      }
      {
        name = "VM_DATA_DIR";
        eval = "$PRJ_DATA_DIR/vms";
      }
    ];

    devshell.startup = {
      setup-test-vms.text = ''
        set -euo pipefail

        [ -d $VM_DATA_DIR ] && exit 0
        mkdir -p $VM_DATA_DIR

        for i in {1..${builtins.toString numAgents}}
        do
          OUT="$VM_DATA_DIR/test-vm-$i"
          mkdir -p $OUT
          ssh-keygen -t ed25519 -q -C root@test-vm-$i -N "" -f "$OUT/ssh_host_ed25519_key"
        done
      '';
    };

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
        ]
        # platform dependent CGO dependencies
        (mkIf isLinux [
          pkgs.gcc
        ])
        (mkIf isDarwin [
          pkgs.darwin.cctools
        ])
      ];

    commands = [
      {
        package = perSystem.flake-linter.default;
      }
      {
        category = "docs";
        package = pkgs.vhs;
        help = "generate terminal gifs";
      }
      {
        category = "docs";
        help = "regenerate gifs for docs";
        name = "gifs";
        command = ''
          set -xeuo pipefail

          for tape in $PRJ_ROOT/docs/vhs/*; do
              vhs $tape -o "$PRJ_ROOT/docs/assets/$(basename $tape .tape).gif"
          done
        '';
      }
      {
        name = "nits";
        category = "development";
        command = ''nix run .#nits -- "$@"'';
      }
      {
        category = "development";
        package = perSystem.gomod2nix.default;
      }
      {
        category = "development";
        package = pkgs.enumer;
      }
      {
        category = "development";
        help = "run local dev services";
        name = "dev";
        command = ''nix run .#dev "$@"'';
      }
      {
        category = "development";
        help = "re-initialise data directory";
        name = "dev-init";
        command = "rm -rf $PRJ_DATA_DIR && direnv reload";
      }
      {
        category = "development";
        help = "run an agent vm";
        name = "run-test-vm";
        command = "nix run .#nixosConfigurations.${system}_test-vm-$1.config.system.build.vmWithBootLoader";
      }
    ];
  }
