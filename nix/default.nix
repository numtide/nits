{inputs, ...}: {
  imports = [
    inputs.flake-root.flakeModule
    ./checks.nix
    ./treefmt.nix
    ./nixos
    ./packages.nix
    ./devshell.nix
    ./dev
  ];
}
