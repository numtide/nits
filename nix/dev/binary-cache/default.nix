{
  perSystem = {
    inputs',
    pkgs,
    ...
  }: let
    secretKey = ./key.sec;
    publicKey = ./key.pub;
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
          export BINARY_CACHE_PUBLIC_KEY=${builtins.readFile publicKey}
        '';
      };
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
