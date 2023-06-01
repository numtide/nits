{
  perSystem = {self', ...}: {
    config.devshells.default = {
        env = [
            { name = "CACHE_DATA_DIR"; eval = "$PRJ_DATA_DIR/cache"; }
        ];
      devshell.startup = {
        setup-cache.text = ''
          [ -d $CACHE_DATA_DIR ] && exit 0
          mkdir -p $CACHE_DATA_DIR
          nix-store --generate-binary-cache-key nits-cache "$CACHE_DATA_DIR/key.sec" "$CACHE_DATA_DIR/key.pub"
        '';
      };
    };

    config.process-compose.configs = {
      dev-services.processes = {
        cache = {
          environment = [
            "LOG_LEVEL=info"
            "LOG_DEVELOPMENT=true"
            "NATS_SEED_FILE=$CACHE_DATA_DIR/user.seed"
            "NATS_JWT_FILE=$CACHE_DATA_DIR/user.jwt"
            "NITS_CACHE_PRIVATE_KEY_FILE=$PRJ_DATA_DIR/cache/key.sec"
          ];
          command = "${self'.packages.nits}/bin/nits cache run";
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
