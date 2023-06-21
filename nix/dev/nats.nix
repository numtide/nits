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
        nats-permissions = {
          command = "nsc push --all";
          depends_on = {
            nats-server.condition = "process_healthy";
          };
        };
      };
    };

    config.devshells.default = {
      devshell.startup = {
        setup-nats = {
          deps = ["setup-agent-vms" "setup-server"];
          text = ''
            set -euo pipefail

            # we only setup the data dir if it doesn't exist
            # to refresh simply delete the directory and run `direnv reload`
            [ -d $NSC_HOME ] && exit 0

            mkdir -p $NSC_HOME

            # initialise nsc state
            nsc init -n numtide --dir $NSC_HOME
            nsc edit operator \
              --service-url nats://localhost:4222 \
              --account-jwt-server-url nats://localhost:4222

            # setup server config
            mkdir -p $NATS_HOME
            cp ${config} "$NATS_HOME/nats.conf"
            nsc generate config --nats-resolver --config-file "$NATS_HOME/auth.conf"

            # generate a sys context
            nsc generate context -a SYS -u sys --context sys

            # enable jetstream for numtide account, no limits
            nsc edit account -n numtide \
              --js-mem-storage -1 \
              --js-disk-storage -1 \
              --js-streams -1 \
              --js-consumer -1

            # set default permissions for numtide account to deny pubsub to anything
            nsc edit account -n numtide --deny-pubsub '>'

            # generate a user for the server and allow it to pubsub everything
            nsc add user -a numtide -n server --allow-pubsub '>'
            nsc export keys --user server --dir "$SERVER_DATA_DIR"
            rm $SERVER_DATA_DIR/O*.nk
            rm $SERVER_DATA_DIR/A*.nk
            mv $(ls $SERVER_DATA_DIR/U*.nk | head) "$SERVER_DATA_DIR/user.seed"

            nsc describe user -n server -R > "$SERVER_DATA_DIR/user.jwt"
            nsc generate context -a numtide -u server --context server

            # generate users for the agent vms
            for AGENT_DIR in $VM_DATA_DIR/*; do
               NKEY=$(${self'.packages.nits}/bin/agent nkey "$AGENT_DIR/ssh_host_ed25519_key")
               BASENAME=$(basename $AGENT_DIR)

               nsc add user -a numtide -k $NKEY -n $BASENAME \
                --allow-pub nits.log.agent.$NKEY \
                --allow-sub nits.inbox.agent.$NKEY.\> \
                --allow-pub \$JS.API.STREAM.INFO.KV_deployment,\$JS.API.CONSUMER.CREATE.KV_deployment \
                --allow-sub \$JS.API.DIRECT.GET.KV_deployment.\$KV.deployment.$NKEY,\$KV.deployment.$NKEY \
                --allow-pub \$JS.API.STREAM.INFO.KV_deployment-result,\$KV.deployment-result.$NKEY \
                --allow-pub \$JS.API.STREAM.INFO.KV_nar-info,\$JS.API.DIRECT.GET.KV_nar-info.\> \
                --allow-pub \$JS.API.STREAM.INFO.OBJ_nar,\$JS.API.STREAM.MSG.GET.OBJ_nar, \
                --allow-pub \$JS.API.STREAM.NAMES,\$JS.API.CONSUMER.CREATE.OBJ_nar,\$JS.FC.OBJ_nar.\> \
                --allow-pub \$JS.API.CONSUMER.DELETE.OBJ_nar.\> \
                --allow-sub \$O.nar.\>

               nsc describe user -n $BASENAME -R > $AGENT_DIR/user.jwt
               echo "$NKEY" > "$AGENT_DIR/nkey.pub"
            done
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
