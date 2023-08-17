{
  lib,
  config,
  pkgs,
  hostname,
  modulesPath,
  ...
}: {
  imports = ["${toString modulesPath}/virtualisation/qemu-vm.nix"];

  nix = {
    nixPath = [
      "nixpkgs=${pkgs.path}"
    ];
    settings = {
      experimental-features = "nix-command flakes";
    };
  };

  networking.hostName = hostname;
  system.stateVersion = config.system.nixos.version;
  boot.loader.grub.devices = lib.mkForce ["/dev/sda"];
  fileSystems."/".device = lib.mkDefault "/dev/sda";

  users.users.root.initialPassword = "";

  services.openssh = {
    enable = true;
    settings = {
      PermitRootLogin = "yes";
    };
  };
}
