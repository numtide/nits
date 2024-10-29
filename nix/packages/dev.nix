{ inputs, perSystem, pkgs, ...}:
(import inputs.process-compose-flake.lib { inherit pkgs; }).makeProcessCompose {
  name = "dev";
  modules = [
    {
      settings.log_location = "$PRJ_DATA_DIR/dev.log";
    }
  ];
}
