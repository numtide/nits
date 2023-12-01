{
  perSystem = {
    inputs',
    pkgs,
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
          environment = let
            config = pkgs.writeText "harmonia-config" ''
              bind = "[::]:3000"
              workers = 1
              max_connection_rate = 256
              priority = 30
            '';
          in [
            "CONFIG_FILE=${config}"
            "SIGN_KEY_PATH=${secretKey}"
          ];
          command = ''
            ${inputs'.harmonia.packages.harmonia}/bin/harmonia
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
