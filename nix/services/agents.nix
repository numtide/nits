{
  self,
  inputs,
  lib,
  ...
}: let
  inherit (inputs) srvos nixpkgs;

  mkAgentHost = {
    id,
    password ? "",
  }:
    nixpkgs.lib.nixosSystem rec {
      system = "x86_64-linux";
      pkgs = import nixpkgs {
        inherit system;
        overlays = [
          self.overlays.default
        ];
      };
      modules = [
        ({
          config,
          pkgs,
          lib,
          modulesPath,
          ...
        }: let
          hostname = "agent-host-${builtins.toString id}";
        in {
          imports = [
            "${toString modulesPath}/virtualisation/qemu-vm.nix"
            self.nixosModules.agent
          ];

          nix = {
            nixPath = [
              "nixpkgs=${pkgs.path}"
            ];
            settings = {
              experimental-features = "nix-command flakes";
              trusted-public-keys = [
                (lib.readFile ./cache/key.pub)
              ];
            };
          };

          networking.hostName = hostname;
          system.stateVersion = config.system.nixos.version;
          boot.loader.grub.devices = lib.mkForce ["/dev/sda"];
          fileSystems."/".device = lib.mkDefault "/dev/sda";

          environment.systemPackages = [
            pkgs.sysz
          ];

          virtualisation = {
            graphics = false;
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

          users.users.root.initialPassword = password;

          services.openssh = {
            enable = true;
            settings = {
              PermitRootLogin = "yes";
            };
          };

          services.nits.agent = {
            logLevel = "debug";
            nats = {
              url = "nats://10.0.2.2";
              jwtFile = "/mnt/shared/user.jwt";
            };
          };
        })
      ];
    };

  numAgents = 1;
in {
  flake.nixosConfigurations = let
    configs =
      map
      (id: lib.nameValuePair "agent-host-${builtins.toString id}" (mkAgentHost {inherit id;}))
      (lib.range 1 numAgents);
  in
    builtins.listToAttrs configs;

  perSystem = {
    pkgs,
    config,
    ...
  }: let
    cfg = config.dev.agents;
  in {
    config.devshells.default = {
      env = [
        {
          name = "VM_DATA_DIR";
          eval = "$PRJ_DATA_DIR/vms";
        }
      ];

      devshell.startup = {
        setup-agent-vms.text = ''
          set -euo pipefail

          [ -d $VM_DATA_DIR ] && exit 0
          mkdir -p $VM_DATA_DIR

          for i in {1..${builtins.toString numAgents}}
          do
            OUT="$VM_DATA_DIR/agent-host-$i"
            mkdir -p $OUT
            ssh-keygen -t ed25519 -q -C root@agent-host-$i -N "" -f "$OUT/ssh_host_ed25519_key"
          done
        '';
      };
    };

    config.process-compose.configs = {
      dev.processes = let
        mkAgentProcess = id: {
          command = "nix run .#nixosConfigurations.agent-host-${builtins.toString id}.config.system.build.vm";
          depends_on = {
            nats-server.condition = "process_healthy";
            nats-permissions.condition = "process_completed";
          };
        };
        configs =
          map
          (id: lib.nameValuePair "agent-host-${builtins.toString id}" (mkAgentProcess id))
          (lib.range 1 numAgents);
      in
        builtins.listToAttrs configs;
    };
  };
}
