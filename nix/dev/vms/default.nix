{
  self,
  lib,
  ...
}: let
  mkTestVM = {
    id,
    self,
    pkgs,
    modules ? [
      ./modules/base.nix
      ./modules/agent.nix
    ],
  }:
    lib.nixosSystem rec {
      inherit pkgs modules;
      specialArgs = {
        inherit self id;
        hostname = "test-vm-${builtins.toString id}";
      };
    };

  numAgents = 3;
in {
  perSystem = {
    pkgs,
    system,
    config,
    lib,
    ...
  }: {
    config.nixosConfigurations = let
      specs = map (id: {
        inherit id;
        name = "test-vm-${toString id}";
      }) (lib.range 1 numAgents);

      configs = builtins.listToAttrs (map ({
        id,
        name,
      }:
        lib.nameValuePair name (mkTestVM {inherit id self pkgs;}))
      specs);

      configsNoHello =
        lib.mapAttrs'
        (
          name: config:
            lib.nameValuePair
            "${name}-no-hello"
            (config.extendModules {modules = [{config.systemd.services.hello.enable = false;}];})
        )
        configs;
    in
      configs // configsNoHello;

    config.devshells.default = {
      env = [
        {
          name = "VM_DATA_DIR";
          eval = "$PRJ_DATA_DIR/vms";
        }
      ];

      devshell.startup = {
        setup-test-vms.text = ''
          set -euo pipefail

          [ -d $VM_DATA_DIR ] && exit 0
          mkdir -p $VM_DATA_DIR

          for i in {1..${builtins.toString numAgents}}
          do
            OUT="$VM_DATA_DIR/test-vm-$i"
            mkdir -p $OUT
            ssh-keygen -t ed25519 -q -C root@test-vm-$i -N "" -f "$OUT/ssh_host_ed25519_key"
          done
        '';
      };

      commands = [
        {
          category = "development";
          help = "run an agent vm";
          name = "run-test-vm";
          command = "nix run .#nixosConfigurations.${system}_test-vm-$1.config.system.build.vmWithBootLoader";
        }
      ];
    };

    config.process-compose = {
      dev.settings.processes = let
        mkAgentProcess = id: {
          command = "run-test-vm ${builtins.toString id}";
          depends_on = {
            binary-cache.condition = "process_healthy";
          };
        };
        configs =
          map
          (id: lib.nameValuePair "test-vm-${builtins.toString id}" (mkAgentProcess id))
          (lib.range 1 numAgents);
      in
        builtins.listToAttrs configs;
    };
  };
}
