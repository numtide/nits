{inputs, ...}: {
  imports = [
    ./agents.nix
    ./guvnor
    ./nats.nix
  ];

  perSystem = {self', ...}: {
    config.devshells.default = {
      commands = [
        {
          category = "development";
          help = "run local dev services";
          package = self'.packages.dev;
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
