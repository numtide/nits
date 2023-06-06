{
  perSystem = {self', ...}: {
    config.devshells.default = {
      env = [
        {
          name = "GUVNOR_DATA_DIR";
          eval = "$PRJ_DATA_DIR/guvnor";
        }
      ];
      devshell.startup = {
        setup-guvnor.text = ''
          [ -d $GUVNOR_DATA_DIR ] && exit 0
          mkdir -p $GUVNOR_DATA_DIR
        '';
      };
    };

    config.process-compose.configs = {
      dev.processes = {
        guvnor = {
          environment = [
            "LOG_LEVEL=info"
            "NATS_SEED_FILE=$GUVNOR_DATA_DIR/user.seed"
            "NATS_JWT_FILE=$GUVNOR_DATA_DIR/user.jwt"
          ];
          command = "${self'.packages.nits}/bin/guvnor run";
          depends_on = {
            nats-server.condition = "process_healthy";
            nats-permissions.condition = "process_completed";
          };
        };
      };
    };
  };
}
