{lib, ...}: {
  perSystem = {pkgs, ...}: {
    config.process-compose = {
      dev.settings.processes = {
        nats-server = {
          working_dir = "$NATS_HOME";
          command = ''${lib.getExe pkgs.nats-server} -c ./nats.conf -sd ./'';
          readiness_probe = {
            http_get = {
              host = "127.0.0.1";
              port = 8222;
              path = "/healthz";
            };
            initial_delay_seconds = 2;
          };
        };
        nits-setup = {
          depends_on = {
            nats-server.condition = "process_started";
          };
          environment = {
            # ensures contexts are generated in the .data directory
            XDG_CONFIG_HOME = "$PRJ_DATA_DIR";
          };
          command = pkgs.writeShellApplication {
            name = "nits-setup";
            runtimeInputs = with pkgs; [jq nsc nits];
            text = ''
              nits cluster add Numtide

              for AGENT_DIR in "$VM_DATA_DIR"/*; do
                 AGENT_NAME=$(basename "$AGENT_DIR")
                 nits agent add --cluster Numtide --private-key-file "$AGENT_DIR/ssh_host_ed25519_key" "$AGENT_NAME"
                 nsc describe user -a Numtide -n "$AGENT_NAME" -R > "$AGENT_DIR/user.jwt"
              done

              # push account changes
              nsc push

              # generate sys context
              nsc generate context -a SYS -u sys --context sys
              nsc generate context -a Numtide -u Admin --context NumtideAdmin
            '';
          };
        };
      };
    };
  };
}
