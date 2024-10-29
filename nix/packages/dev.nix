{ inputs, perSystem, pkgs, ...}:
let
  inherit (pkgs) lib;
in
(import inputs.process-compose-flake.lib { inherit pkgs; }).makeProcessCompose {
  name = "dev";
  modules = [
    {
      settings.log_location = "$PRJ_DATA_DIR/dev.log";
    }

    # START VMS
    {
      settings.processes = let
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
    }
  ];
}
