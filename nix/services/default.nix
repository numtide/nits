{
  inputs,
  lib,
  ...
}: {
  perSystem = {
    system,
    self',
    pkgs,
    ...
  }: {
    config.process-compose.configs = {
      dev-services.processes = let
        config = ./nats.conf;
      in {
        nats.command = "${lib.getExe pkgs.nats-server} -c ${config}";
      };
    };

    config.devshells.default = {
      commands = let
        category = "development";
      in [
        {
          inherit category;
          help = "run local dev services";
          package = self'.packages.dev-services;
        }
      ];
    };
  };
}
