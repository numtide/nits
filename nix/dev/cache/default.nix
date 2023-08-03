{
  perSystem = {self', ...}: {
    config.devshells.default = {
      env = [
        {
          name = "CACHE_DATA_DIR";
          eval = "$PRJ_DATA_DIR/cache";
        }
        {
          name = "CACHE_URL";
          eval = "http://localhost:3000";
        }
        {
          name = "CACHE_COPY_URL";
          eval = "$CACHE_URL/\?compression\=zstd";
        }
      ];
      devshell.startup = {
        setup-server.text = ''
          [ -d $CACHE_DATA_DIR ] && exit 0
          mkdir -p $CACHE_DATA_DIR
        '';
      };

      commands = [
        {
          category = "development";
          help = "copy a store path to the binary cache";
          name = "copy-to-cache";
          # refresh flag is important otherwise nix will use ~/.cache/nix to avoid sending paths it thinks are
          # already in the cache
          command = "nix copy -v --refresh --to $CACHE_COPY_URL $1";
        }
      ];
    };

    config.process-compose = {
      dev.settings.processes = {
        nits-cache = {
          environment = let
            keyFile = ./key.sec;
          in [
            "LOG_LEVEL=info"
            "NATS_SEED_FILE=$CACHE_DATA_DIR/user.seed"
            "NATS_JWT_FILE=$CACHE_DATA_DIR/user.jwt"
            "NITS_CACHE_PRIVATE_KEY_FILE=${keyFile}"
          ];
          command = "${self'.packages.nits}/bin/nits-cache";
          depends_on = {
            nats-server.condition = "process_healthy";
            nats-permissions.condition = "process_completed";
          };
          readiness_probe = {
            http_get = {
              host = "127.0.0.1";
              port = 3000;
              path = "/nix-cache-info";
            };
            initial_delay_seconds = 2;
          };
        };
      };
    };
  };
}
