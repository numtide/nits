{inputs, ...}: {
  imports = [
    inputs.flake-root.flakeModule
    ./checks.nix
    ./formatter.nix
    ./packages.nix
    ./shell.nix
    ./services
  ];
}
