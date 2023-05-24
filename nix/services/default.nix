{inputs, ...}: {
  imports = [
    ./cache.nix
    ./nats.nix
  ];

  perSystem = {self', ...}: {
    config.devshells.default = {
      commands = [
        {
          category = "development";
          help = "run local dev services";
          package = self'.packages.dev-services;
        }
      ];
    };
  };
}
