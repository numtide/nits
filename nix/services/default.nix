{inputs, ...}: {
  imports = [
    ./agents.nix
    ./cache.nix
    ./guvnor.nix
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
      ];
    };
  };
}
