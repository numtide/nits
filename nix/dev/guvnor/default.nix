{
  perSystem = {self', ...}: {
    config.devshells.default = {
      env = [
        {
          name = "GUVNOR_DATA_DIR";
          eval = "$PRJ_DATA_DIR/guvnor";
        }
        {
          name = "GUVNOR_URL";
          eval = "http://localhost:3000";
        }
        {
          name = "GUVNOR_CACHE_URL";
          eval = "$GUVNOR_URL/\?compression\=zstd";
        }
      ];
      devshell.startup = {
        setup-guvnor.text = ''
          [ -d $GUVNOR_DATA_DIR ] && exit 0
          mkdir -p $GUVNOR_DATA_DIR
        '';
      };

      commands = [
        {
          category = "development";
          help = "copy a store path to the guvnor binary cache";
          name = "copy-to-guvnor";
          # refresh flag is important otherwise nix will use ~/.cache/nix to avoid sending paths it thinks are
          # already in the cache
          command = "nix copy -v --refresh --to $GUVNOR_CACHE_URL $1";
        }
      ];
    };

    config.process-compose = {
      dev.settings.processes = {
        guvnor = {
          environment = let
            keyFile = ./key.sec;
          in [
            "LOG_LEVEL=info"
            "NATS_SEED_FILE=$GUVNOR_DATA_DIR/user.seed"
            "NATS_JWT_FILE=$GUVNOR_DATA_DIR/user.jwt"
            "NITS_CACHE_PRIVATE_KEY_FILE=${keyFile}"
          ];
          command = "${self'.packages.nits}/bin/guvnor run";
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
