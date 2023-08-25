{inputs, ...}: {
  imports = [
    inputs.flake-root.flakeModule
    ./checks.nix
    ./formatter.nix
    ./nixos
    ./packages.nix
    ./devshell.nix
    ./dev
  ];
}
