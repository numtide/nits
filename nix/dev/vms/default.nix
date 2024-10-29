{
  self,
  lib,
  ...
}: let
  mkTestVM = {
    id,
    self,
    pkgs,
    modules ? [
      ./modules/base.nix
      ./modules/agent.nix
    ],
  }:
    lib.nixosSystem rec {
      inherit pkgs modules;
      specialArgs = {
        inherit self id;
        hostname = "test-vm-${builtins.toString id}";
      };
    };

  numAgents = 3;
in {
  perSystem = {
    pkgs,
    system,
    config,
    lib,
    ...
  }: {
    config.nixosConfigurations = let
      specs = map (id: {
        inherit id;
        name = "test-vm-${toString id}";
      }) (lib.range 1 numAgents);

      configs = builtins.listToAttrs (map ({
        id,
        name,
      }:
        lib.nameValuePair name (mkTestVM {inherit id self pkgs;}))
      specs);

      configsNoHello =
        lib.mapAttrs'
        (
          name: config:
            lib.nameValuePair
            "${name}-no-hello"
            (config.extendModules {modules = [{config.systemd.services.hello.enable = false;}];})
        )
        configs;
    in
      configs // configsNoHello;
  };
}
