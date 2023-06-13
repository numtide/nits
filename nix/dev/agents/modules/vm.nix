{
  id,
  hostname,
  ...
}: {
  virtualisation = {
    graphics = false;
    diskSize = 5120;
    diskImage = "$VM_DATA_DIR/${hostname}/disk.qcow2";

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
      rm /etc/ssh/ssh_host_ed25519_key*
      cp /mnt/shared/ssh_host_ed25519_key /etc/ssh/ssh_host_ed25519_key
      cp /mnt/shared/ssh_host_ed25519_key.pub /etc/ssh/ssh_host_ed25519_key.pub

      chmod 600 /etc/ssh/ssh_host_ed25519_key
      chmod 644 /etc/ssh/ssh_host_ed25519_key.pub
    '';
  };
}
