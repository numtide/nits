{
  lib,
  pkgs,
  config,
  ...
}: let
  cfg = config.services.nits.agent;
in {
  options.services.nits.agent = with lib; {
    nats = {
      url = mkOption {
        type = types.str;
        description = "NATS server url.";
      };
      jwtFile = mkOption {
        type = types.path;
        description = "Path to a file containing a NATS JWT token.";
      };
      hostKeyFile = mkOption {
        type = types.path;
        description = "Path to an ed25519 host key file";
        default = "/etc/ssh/ssh_host_ed25519_key";
      };
    };
  };

  config = {
    systemd.services.nits-agent = {
      after = ["network.target"];
      wantedBy = ["sysinit.target"];

      description = "Nits Agent";

      startLimitIntervalSec = 0;

      environment = {
        NATS_URL = cfg.nats.url;
        LOG_LEVEL = "debug";
      };

      serviceConfig = with lib; {
        Restart = mkDefault "on-failure";
        RestartSec = 1;

        User = "nits-agent";
        StateDirectory = "nits-agent";
        ExecStart = "${pkgs.nits}/bin/nits agent --nats-host-key-file %d/host_key --nats-jwt-file %d/jwt";

        LoadCredential = [
          "jwt:${cfg.nats.jwtFile}"
          "host_key:${cfg.nats.hostKeyFile}"
        ];

        # https://www.freedesktop.org/software/systemd/man/systemd.exec.html#DynamicUser=
        # Enabling dynamic user implies other options which cannot be changed:
        #   * RemoveIPC = true
        #   * PrivateTmp = true
        #   * NoNewPrivileges = "strict"
        #   * RestrictSUIDSGID = true
        #   * ProtectSystem = "strict"
        #   * ProtectHome = "read-only"
        DynamicUser = mkDefault true;

        ProtectClock = mkDefault true;
        ProtectProc = mkDefault "noaccess";
        ProtectKernelLogs = mkDefault true;
        ProtectKernelModules = mkDefault true;
        ProtectKernelTunables = mkDefault true;
        ProtectControlGroups = mkDefault true;
        ProtectHostname = mkDefault true;
        PrivateDevices = mkDefault true;
        RestrictRealtime = mkDefault true;
        RestrictNamespaces = mkDefault true;
        LockPersonality = mkDefault true;
        MemoryDenyWriteExecute = mkDefault true;
        SystemCallFilter = mkDefault ["@system-service" "~@privileged"];
      };
    };
  };
}
