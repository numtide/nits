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
      {
        name = "NATS_HOME";
        eval = "$PRJ_DATA_DIR/nats";
      }
      {
        name = "NSC_HOME";
        eval = "$PRJ_DATA_DIR/nsc";
      }
      {
        name = "NKEYS_PATH";
        eval = "$NSC_HOME";
      }
      {
        name = "NATS_JWT_DIR";
        eval = "$PRJ_DATA_DIR/nats/jwt";
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

      setup-nats = let
        config = pkgs.writeTextFile {
          name = "nats.conf";
          text = ''
            ## Default NATS server configuration (see: https://docs.nats.io/running-a-nats-service/configuration)

            ## Host for client connections.
            host: "127.0.0.1"

            ## Port for client connections.
            port: 4222

            ## Port for monitoring
            http_port: 8222

            ## Configuration map for JetStream.
            ## see: https://docs.nats.io/running-a-nats-service/configuration#jetstream
            jetstream {}

            # include paths must be relative so for simplicity we just read in the auth.conf file
            include './auth.conf'
          '';
        };
      in {
        deps = ["setup-test-vms"];
        text = ''
          set -euo pipefail

          # we only setup the data dir if it doesn't exist
          # to refresh simply delete the directory and run `direnv reload`
          [ -d $NSC_HOME ] && exit 0

          mkdir -p $NSC_HOME

          # initialise nsc state

          nsc init -n Nits --dir $NSC_HOME
          nsc edit operator \
              --service-url nats://localhost:4222 \
              --account-jwt-server-url nats://localhost:4222

          # setup server config

          mkdir -p $NATS_HOME
          cp ${config} "$NATS_HOME/nats.conf"
          nsc generate config --nats-resolver --config-file "$NATS_HOME/auth.conf"
        '';
      };
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
        # nats
        [
          pkgs.nkeys
          pkgs.nats-top
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
      {
        category = "nats";
        name = "nsc";
        command = ''XDG_CONFIG_HOME=$PRJ_DATA_DIR ${pkgs.nsc}/bin/nsc -H "$NSC_HOME" "$@"'';
      }
      {
        category = "nats";
        name = "nats";
        command = ''XDG_CONFIG_HOME=$PRJ_DATA_DIR ${pkgs.natscli}/bin/nats "$@"'';
      }
    ];
  }
