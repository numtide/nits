{
  inputs,
  lib,
  ...
}: {
  perSystem = {
    system,
    self',
    pkgs,
    ...
  }: let
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
    config.process-compose = {
      dev.settings.processes = {
        nats-server = {
          working_dir = "$NATS_HOME";
          command = ''${lib.getExe pkgs.nats-server} -c ./nats.conf -sd ./'';
          readiness_probe = {
            http_get = {
              host = "127.0.0.1";
              port = 8222;
              path = "/healthz";
            };
            initial_delay_seconds = 2;
          };
        };
        nits-setup = {
          depends_on = {
            nats-server.condition = "process_started";
          };
          environment = {
            # ensures contexts are generated in the .data directory
            XDG_CONFIG_HOME = "$PRJ_DATA_DIR";
          };
          command = pkgs.writeShellApplication {
            name = "nats-setup";
            runtimeInputs = [pkgs.jq pkgs.nsc self'.packages.nits];
            text = ''
              NITS_PK=$(nsc describe account -n Nits --raw -J | jq -r .sub)
              nits add cache --account Nits
              nits add cluster --name Numtide --nits-public-key "$NITS_PK"

              for AGENT_DIR in "$VM_DATA_DIR"/*; do
                 AGENT_NAME=$(basename "$AGENT_DIR")
                 nits add agent --cluster Numtide --name "$AGENT_NAME" --private-key-file "$AGENT_DIR/ssh_host_ed25519_key"
                 nsc describe user -a Numtide -n "$AGENT_NAME" -R > "$AGENT_DIR/user.jwt"
              done

              # generate sys context
              nsc generate context -a SYS -u sys --context sys

              # export credentials for cache
              nsc export keys --account Nits --user Cache --dir "$CACHE_DATA_DIR" --include-jwts

              find "$PRJ_DATA_DIR/cache" -type f \
                  -regextype posix-extended -regex '.*\/[OA].*\.(nk|jwt)$' \
                  -exec rm {} \;

              find "$CACHE_DATA_DIR"/U*.nk -type f -regex '.*\.nk$' -exec mv {} "$CACHE_DATA_DIR/user.seed" \;
              find "$CACHE_DATA_DIR"/U*.jwt -type f -regex '.*\.jwt$' -exec mv {} "$CACHE_DATA_DIR/user.jwt" \;
            '';
          };
        };
      };
    };

    config.devshells.default = {
      devshell.startup = {
        setup-nats = let
          nits = lib.getExe self'.packages.default;
        in {
          deps = ["setup-agent-vms" "setup-server"];
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

      env = [
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

      packages = [
        pkgs.nkeys
        pkgs.nats-top
      ];

      commands = let
        category = "nats";
      in [
        {
          inherit category;
          name = "nsc";
          command = ''XDG_CONFIG_HOME=$PRJ_DATA_DIR ${pkgs.nsc}/bin/nsc -H "$NSC_HOME" "$@"'';
        }
        {
          inherit category;
          name = "nats";
          command = ''XDG_CONFIG_HOME=$PRJ_DATA_DIR ${pkgs.natscli}/bin/nats "$@"'';
        }
      ];
    };
  };
}
