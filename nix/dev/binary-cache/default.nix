{
  perSystem = {
    inputs',
    pkgs,
    lib,
    ...
  }: let
    secretKey = ./key.sec;
  in {
    config.devshells.default = {
      env = [
        {
          name = "BINARY_CACHE_DATA";
          eval = "$PRJ_DATA_DIR/binary-cache";
        }
        {
          name = "BINARY_CACHE_PORT";
          value = "3000";
        }
      ];

      devshell.startup = {
        export-public-key.text = ''
          export BINARY_CACHE_PUBLIC_KEY=$(nix key convert-secret-to-public < $PRJ_ROOT/nix/dev/binary-cache/key.sec)
        '';
      };

      commands = [
        {
          category = "dev";
          help = "Recursively sign a given store path with the local dev binary cache key";
          package = pkgs.writeShellApplication {
            name = "sign-path";
            text = ''nix store sign --key-file ${secretKey} --recursive "$1"'';
          };
        }
      ];
    };

    config.process-compose = {
      dev.settings.processes = {
        binary-cache = {
          environment = [
            "NIX_SECRET_KEY_FILE=${secretKey}"
          ];
          command = ''
            ${inputs'.nix-serve.packages.nix-serve}/bin/nix-serve \
                --host 127.0.0.1 \
                --port "$BINARY_CACHE_PORT" \
                --access-log /dev/stderr
          '';
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
