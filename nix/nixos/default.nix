{
  config,
  inputs,
  ...
}: {
  imports = [
    inputs.nix-lib.flakeModules.nixos
  ];

  config.flake.nixosModules = {
    agent = import ./modules/agent.nix;
  };
}
