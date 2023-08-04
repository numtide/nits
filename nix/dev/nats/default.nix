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
        nits-numtide-config = {
          depends_on = {
            nats-permissions.condition = "process_completed_successfully";
          };
          command = pkgs.writeShellScriptBin "nits-numtide-config" ''
            nats --context numtide-admin kv add deployment \
                --history 64 \
                --republish-source '$KV.deployment.*' \
                --republish-destination 'NITS.AGENT.{{wildcard(1)}}.DEPLOYMENT'

            nats --context numtide-admin kv add deployment-result --history 64

            nats --context numtide-admin stream add --config ${./agent-logs.json}
            nats --context numtide-admin stream add --config ${./agent-deployments.json}
          '';
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
            nsc init -n nits --dir $NSC_HOME
            nsc edit operator \
              --service-url nats://localhost:4222 \
              --account-jwt-server-url nats://localhost:4222

            NITS_PK=$(nsc describe account -n nits --raw -J | jq -r .sub)

            # setup server config
            mkdir -p $NATS_HOME
            cp ${config} "$NATS_HOME/nats.conf"
            nsc generate config --nats-resolver --config-file "$NATS_HOME/auth.conf"

            # generate a sys context
            nsc generate context -a SYS -u sys --context sys

            # enable jetstream for nits account, no limits
            nsc edit account -n nits \
              --js-mem-storage -1 \
              --js-disk-storage -1 \
              --js-streams -1 \
              --js-consumer -1

            # generate a user for the nits cache
            nsc add user -a nits -n cache
            nsc export keys --user cache --dir "$CACHE_DATA_DIR"

            rm $CACHE_DATA_DIR/O*.nk
            rm $CACHE_DATA_DIR/A*.nk
            mv $(ls $CACHE_DATA_DIR/U*.nk | head) "$CACHE_DATA_DIR/user.seed"

            nsc describe user -a nits -n cache -R > "$CACHE_DATA_DIR/user.jwt"
            nsc generate context -a nits -u cache --context cache

            # export cache service
            nsc add export -a nits --private \
                --name "Binary Cache" \
                --subject "NITS.CACHE.>" \
                --service --response-type Chunked

            # generate a numtide account for our cluster
            nsc add account -n numtide --deny-pubsub '>'
            NUMTIDE_PK=$(nsc describe account -n numtide --raw -J | jq -r .sub)

            # enable jetstream, no limits
            nsc edit account -n numtide \
                --js-mem-storage -1 \
                --js-disk-storage -1 \
                --js-streams -1 \
                --js-consumer -1

            # import cache service
            TOKEN="$(mktemp -d)/numtide"
            nsc generate activation \
                --account nits --subject 'NITS.CACHE.>' \
                --output-file "$TOKEN" --target-account "$NUMTIDE_PK"
            nsc add import -a numtide -n "Binary Cache" --token "$TOKEN"

            # add admin user
            nsc add user -a numtide -n admin --allow-pubsub '>'
            nsc generate context -a numtide -u admin --context numtide-admin

            # generate users for the agent vms
            for AGENT_DIR in $VM_DATA_DIR/*; do
               NKEY=$(${self'.packages.nits}/bin/nits-agent nkey "$AGENT_DIR/ssh_host_ed25519_key")
               BASENAME=$(basename $AGENT_DIR)

               nsc add user -a numtide -k $NKEY -n $BASENAME \
                --allow-pub NITS.CACHE.\> \
                --allow-pubsub NITS.AGENT.$NKEY.\> \
                --allow-pub \$JS.API.STREAM.NAMES \
                --allow-pub \$JS.API.CONSUMER.\*.agent-deployments.\> \
                --allow-pub \$JS.ACK.agent-deployments.\>

               nsc describe user -a numtide -n $BASENAME -R > $AGENT_DIR/user.jwt
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
