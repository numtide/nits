{
  flake.nixosModules = {
    agent = import ./modules/agent.nix;
  };
}
