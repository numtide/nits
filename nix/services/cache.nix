{
  perSystem = {self', ...}: {
    config.devshells.default = {
      devshell.startup = {
        generate-binary-cache-key.text = ''
          OUT="$PRJ_DATA_DIR/cache"
          [ -d $OUT ] && exit 0
          mkdir -p $OUT
          nix-store --generate-binary-cache-key nits-cache "$OUT/key.sec" "$OUT/key.pub"
        '';
      };
    };

    config.process-compose.configs = {
      dev-services.processes = {
        cache = {
          environment = [
            "LOG_LEVEL=info"
            "LOG_DEVELOPMENT=true"
            "NATS_CREDENTIALS_FILE=$NSC_HOME/creds/numtide/numtide/cache.creds"
            "NITS_CACHE_PRIVATE_KEY_FILE=$PRJ_DATA_DIR/cache/key.sec"
          ];
          command = "${self'.packages.nits}/bin/nits cache";
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
