{...}: {
  imports = [
    ./binary-cache
    ./nats
    ./vms
  ];

  perSystem = _: {
    config.process-compose.dev.settings = {
      log_location = "$PRJ_DATA_DIR/dev.log";
    };

    config.devshells.default = {
      commands = [
        {
          category = "development";
          help = "run local dev services";
          name = "dev";
          command = ''nix run .#dev "$@"'';
        }
        {
          category = "development";
          help = "re-initialise data directory";
          name = "dev-init";
          command = "rm -rf $PRJ_DATA_DIR && direnv reload";
        }
      ];
    };
  };
}
