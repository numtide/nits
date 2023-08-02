{
  perSystem = {self', ...}: {
    config.devshells.default = {
      env = [
        {
          name = "SERVER_DATA_DIR";
          eval = "$PRJ_DATA_DIR/server";
        }
        {
          name = "SERVER_URL";
          eval = "http://localhost:3000";
        }
        {
          name = "SERVER_CACHE_URL";
          eval = "$SERVER_URL/\?compression\=zstd";
        }
      ];
      devshell.startup = {
        setup-server.text = ''
          [ -d $SERVER_DATA_DIR ] && exit 0
          mkdir -p $SERVER_DATA_DIR
        '';
      };

      commands = [
        {
          category = "development";
          help = "copy a store path to the binary cache";
          name = "copy-to-server";
          # refresh flag is important otherwise nix will use ~/.cache/nix to avoid sending paths it thinks are
          # already in the cache
          command = "nix copy -v --refresh --to $SERVER_CACHE_URL $1";
        }
      ];
    };

    config.process-compose = {
      dev.settings.processes = {
        server = {
          environment = let
            keyFile = ./key.sec;
          in [
            "LOG_LEVEL=info"
#            "NATS_SERVER_CONFIG=$NATS_HOME/nats.conf"
            "NATS_SEED_FILE=$SERVER_DATA_DIR/user.seed"
            "NATS_JWT_FILE=$SERVER_DATA_DIR/user.jwt"
            "NITS_CACHE_PRIVATE_KEY_FILE=${keyFile}"
          ];
          command = "${self'.packages.nits}/bin/nits-server";
          depends_on = {
          };
          readiness_probe = {
            http_get = {
              host = "127.0.0.1";
              port = 8222;
              path = "/healthz";
            };
            initial_delay_seconds = 2;
          };
        };
      };
    };
  };
}
