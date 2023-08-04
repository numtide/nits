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

  mkAgentHost = {
    id,
    self,
    pkgs,
    modules ? [
      ./modules/base.nix
      ./modules/vm.nix
      ./modules/agent.nix
    ],
    extraModules ? [],
  }:
    lib.nixosSystem rec {
      inherit pkgs modules;
      inherit (pkgs) system;
      specialArgs = {
        inherit self id;
        hostname = "agent-host-${builtins.toString id}";
      };
    };

  numAgents = 1;
in {
  flake.nixosConfigurations = let
    configs =
      map
      (id: lib.nameValuePair "agent-host-${builtins.toString id}" (mkAgentHost {inherit id self pkgs;}))
      (lib.range 1 numAgents);
  in
    builtins.listToAttrs configs;

  perSystem = {
    pkgs,
    config,
    lib,
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
        {
          category = "development";
          help = "deploy changes to an agent host";
          name = "deploy-agent";
          command = let
            flakeRoot = lib.getExe config.flake-root.package;
            perl = lib.getExe pkgs.perl;
          in ''
            set -euo pipefail

            PATH=$PATH:${perl}

            ID=$1
            ACTION=$2
            CONFIG=$3

            exec 3>&1
            exec 4>&2

            prefix_out () {
                exec 1> >( ${perl} -ne '$| = 1; print "'"[$1]"' | $_"' >&3)
                exec 2> >( ${perl} -ne '$| = 1; print "'"[$1]"' | $_"' >&4)
            }

            prefix_out "build-closure"

            create_derivation () {
                nix-instantiate \
                    --expr '({ flakeRoot, id, mod }: ((builtins.getFlake "path:''${flakeRoot}").nixosConfigurations."agent-host-''${id}".extendModules { modules = [mod];}).config.system.build.toplevel)' \
                    --argstr flakeRoot $PWD \
                    --argstr id $ID \
                    --arg mod "$CONFIG"
            }

            DRV=$(create_derivation)
            STORE_PATH=$(nix-store --realise $DRV)

            prefix_out "copy-to-cache"

            copy-to-cache $STORE_PATH

            prefix_out "update-deployment"

            NKEY=$(cat $VM_DATA_DIR/agent-host-$ID/nkey.pub)
            nats --context numtide-admin kv put deployment $NKEY "{\"action\":\"$ACTION\",\"closure\":\"$STORE_PATH\"}"

            prefix_out "agent-logs"

            nats --context numtide-admin subscribe --stream agent-logs --last "NITS.AGENT.$NKEY.LOGS" --raw
          '';
        }
      ];
    };

    config.process-compose = {
      dev.settings.processes = let
        mkAgentProcess = id: {
          command = "run-agent ${builtins.toString id}";
          depends_on = {
            nits-cache.condition = "process_healthy";
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
