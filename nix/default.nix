{inputs, ...}: {
  imports = [
    inputs.flake-root.flakeModule
    ./checks.nix
    ./treefmt.nix
    ./nixos
    ./nixpkgs.nix
    ./packages.nix
    ./devshell.nix
    ./dev
  ];
}
