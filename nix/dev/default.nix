{...}: {
  imports = [
    ./binary-cache
    ./nats
    ./vms
  ];

  perSystem = _: {
    config.process-compose.dev.settings = {
      log_location = "$PRJ_DATA_DIR/dev.log";
    };
  };
}
