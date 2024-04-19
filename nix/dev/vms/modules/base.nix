{
  lib,
  config,
  pkgs,
  id,
  hostname,
  ...
}: {
  nix = {
    nixPath = [
      "nixpkgs=${pkgs.path}"
    ];
    settings = {
      experimental-features = "nix-command flakes";
      substituters = ["http://10.0.2.2:3000"];
      trusted-public-keys = [(lib.readFile ../../binary-cache/key.pub)];
    };
  };

  networking.hostName = hostname;
  system.stateVersion = config.system.nixos.version;

  boot = {
    growPartition = true;
    kernelParams = ["console=ttyS0"];
    loader = {
      timeout = 0;
      grub.device = "/dev/vda";
    };
  };

  fileSystems."/" = {
    device = "/dev/vda";
    autoResize = true;
    fsType = "ext4";
  };

  users.users.nits = {
    isNormalUser = true;
    extraGroups = ["wheel"];
  };
  services.getty.autologinUser = "nits";
  security.sudo.wheelNeedsPassword = false;

  services.openssh.enable = true;

  virtualisation.vmVariantWithBootLoader = {
    virtualisation = {
      graphics = false;
      diskSize = 5120;
      diskImage = "$VM_DATA_DIR/${hostname}/disk.qcow2";

      writableStoreUseTmpfs = false;

      forwardPorts = [
        {
          from = "host";
          # start at 2222 and increment
          host.port = 2221 + id;
          guest.port = 22;
        }
      ];

      sharedDirectories = {
        config = {
          source = "$VM_DATA_DIR/${hostname}";
          target = "/mnt/shared";
        };
      };
    };

    system.activationScripts = {
      # replace host key with pre-generated one
      host-key.text = ''
        rm -f /etc/ssh/ssh_host_ed25519_key*
        cp /mnt/shared/ssh_host_ed25519_key /etc/ssh/ssh_host_ed25519_key
        cp /mnt/shared/ssh_host_ed25519_key.pub /etc/ssh/ssh_host_ed25519_key.pub

        chmod 600 /etc/ssh/ssh_host_ed25519_key
        chmod 644 /etc/ssh/ssh_host_ed25519_key.pub
      '';
    };
  };
}
