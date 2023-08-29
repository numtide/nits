{inputs, ...}: {
  imports = [
    inputs.flake-parts.flakeModules.easyOverlay
  ];

  perSystem = {
    lib,
    pkgs,
    self',
    ...
  }: {
    packages = rec {
      nits = pkgs.buildGoModule rec {
        pname = "nits";
        version = "0.0.1+dev";

        runtimeInputs = with pkgs; [nsc natscli];

        src = lib.cleanSourceAndNix ../.;
        vendorSha256 = "sha256-fc+YbBBXil36wUg4UqWBRZEGAZrQqzKBfU6lq5TEPcE=";

        ldflags = [
          "-X 'build.Name=${pname}'"
          "-X 'build.Version=${version}'"
        ];

        meta = with lib; {
          description = "Nix & NATS";
          homepage = "https://github.com/numtide/nits";
          license = licenses.apsl20;
          mainProgram = "nits";
        };
      };

      default = nits;
    };

    overlayAttrs = self'.packages;
  };
}
