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
    config.process-compose.configs = {
      dev-services.processes = {
        nats.command = ''
          # start the server
          ${lib.getExe pkgs.nats-server} -c $NATS_HOME/nats.conf -sd $NATS_HOME &

          # push latest account and user permissions
          nsc push --all

          wait  # wait for server to exit
        '';
      };
    };

    config.devshells.default = {
      devshell.startup = {
        setup-nsc.text = ''
          set -euo pipefail

          # we only setup the data dir if it doesn't exist
          # to refresh simply delete the directory and run `direnv reload`
          [ -d $PRJ_DATA_DIR ] && exit 0

          mkdir -p $NATS_HOME $NSC_HOME

          # initialise nsc state
          nsc init -n numtide --dir $NSC_HOME
          nsc edit operator \
            --service-url nats://localhost:4222 \
            --account-jwt-server-url nats://localhost:4222

          # setup server config
          cp ${config} "$NATS_HOME/nats.conf"
          nsc generate config --nats-resolver --config-file "$NATS_HOME/auth.conf"

          # enable jetstream for numtide account, no limits
          nsc edit account -n numtide \
            --js-mem-storage -1 \
            --js-disk-storage -1 \
            --js-streams -1 \
            --js-consumer -1

          # generate some users
          nsc add user -a numtide -n cache

          # generate some cli contexts
          nsc generate context -a numtide -u cache --context cache
        '';
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
        category = "development";
      in [
        {
          inherit category;
          help = "run local dev services";
          package = self'.packages.dev-services;
        }
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
