{
  self,
  inputs,
  ...
}: {
  perSystem = {system, ...}: {
    # customize nixpkgs instance
    _module.args.pkgs = import inputs.nixpkgs {
      inherit system;
      overlays = [
        # required for building agent nixos configs
        self.overlays.default
        # adds buildGoApplication
        inputs.gomod2nix.overlays.default
      ];
    };
  };
}
