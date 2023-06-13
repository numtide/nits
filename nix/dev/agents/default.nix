{
  self,
  inputs,
  lib,
  ...
}: let
  inherit (inputs) srvos nixpkgs;

  pkgs = import nixpkgs {
    system = "x86_64-linux";
    overlays = [
      self.overlays.default
    ];
  };

  numAgents = 1;
in {
  flake.nixosConfigurations = let
    configs =
      map
      (id: lib.nameValuePair "agent-host-${builtins.toString id}" (lib.mkAgentHost {inherit id self pkgs;}))
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

      commands = [
        {
          category = "development";
          help = "run an agent vm";
          name = "run-agent";
          command = "nix run .#nixosConfigurations.agent-host-$1.config.system.build.vm";
        }
        {
          category = "development";
          help = "build an agent vm";
          name = "build-agent";
          command = "nix build .#nixosConfigurations.agent-host-$1.config.system.build.vm";
        }
      ];
    };

    config.process-compose = {
      dev.settings.processes = let
        mkAgentProcess = id: {
          command = "run-agent ${builtins.toString id}";
          depends_on = {
            guvnor.condition = "process_healthy";
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
