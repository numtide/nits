{
  lib,
  pkgs,
  config,
  ...
}: let
  cfg = config.services.nits.agent;
in {
  options.services.nits.agent = with lib; {
    package = mkOption {
      type = types.package;
      default = pkgs.nits;
      defaultText = literalExpression "pkgs.nits";
      description = mdDoc "Package to use for nits.";
    };
    nats = {
      url = mkOption {
        type = types.str;
        example = "nats://localhost:4222";
        description = mdDoc "NATS server url.";
      };
      jwtFile = mkOption {
        type = types.path;
        example = "/mnt/shared/user.jwt";
        description = mdDoc "Path to a file containing a NATS JWT token.";
      };
      hostKeyFile = mkOption {
        type = types.path;
        default = "/etc/ssh/ssh_host_ed25519_key";
        example = "/etc/ssh/ssh_host_ed25519_key";
        description = mdDoc "Path to an ed25519 host key file";
      };
    };
    cacheProxy = {
      subject = mkOption {
        type = types.nullOr types.str;
        default = null;
        example = "NITS.CACHE";
        description = mdDoc "Nats subject for communicating with the Nits Binary Cache";
      };
      publicKey = mkOption {
        type = types.nullOr types.str;
        default = null;
        example = "nits-cache:XaEVMQLJh2gx51XXQ2CaREQSwJ/b6DSNN4bO1bW9nA4=";
        description = mdDoc "Signing key for the Nits Binary Cache";
      };
    };

    logLevel = mkOption {
      type = types.enum ["debug" "info" "warn" "error"];
      default = "info";
      example = "debug";
      description = mdDoc "Selects the logging level.";
    };
  };

  config = {
    systemd.services.nits-agent = {
      after = ["network.target"];
      wantedBy = ["sysinit.target"];

      description = "Nits Agent";
      startLimitIntervalSec = 0;

      # the agent will restart itself after a successful deployment
      restartIfChanged = false;

      path = [
        pkgs.nix
        pkgs.nixos-rebuild
      ];

      environment = lib.filterAttrs (_: v: v != null) {
        NATS_URL = cfg.nats.url;
        NATS_HOST_KEY_FILE = cfg.nats.hostKeyFile;
        NATS_JWT_FILE = cfg.nats.jwtFile;
        LOG_LEVEL = cfg.logLevel;
        NITS_CACHE_PROXY_SUBJECT = cfg.cacheProxy.subject;
        NITS_CACHE_PROXY_PUBLIC_KEY = cfg.cacheProxy.publicKey;
      };

      serviceConfig = with lib; {
        Restart = mkDefault "on-failure";
        RestartSec = 1;

        User = "root";
        StateDirectory = "nits-agent";
        ExecStart = "${cfg.package}/bin/nits-agent";
      };
    };
  };
}
