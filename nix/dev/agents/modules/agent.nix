{
  self,
  pkgs,
  ...
}: {
  imports = [
    self.nixosModules.agent
  ];

  services.nits.agent = {
    logLevel = "debug";
    nats = {
      url = "nats://10.0.2.2";
      jwtFile = "/mnt/shared/user.jwt";
    };
  };

  systemd.services.hello = {
    enable = true;
    after = ["network.target"];
    wantedBy = ["sysinit.target"];
    description = "A test service";

    startLimitIntervalSec = 0;

    serviceConfig = {
      Type = "simple";
      ExecStart = "${pkgs.hello}/bin/hello";
    };
  };
}
