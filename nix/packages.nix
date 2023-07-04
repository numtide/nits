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

        src = lib.cleanSourceAndNix ../.;
        vendorSha256 = "sha256-16gy2XkIjO5pLMQJAebCpVCVF9E3uLMaWj0zwXR/oDY=";

        ldflags = [
          "-X 'build.Name=${pname}'"
          "-X 'build.Version=${version}'"
        ];

        meta = with lib; {
          description = "Nix & NATS";
          homepage = "https://github.com/numtide/nits";
          license = licenses.apsl20;
        };
      };

      default = nits;
    };

    overlayAttrs = self'.packages;
  };
}
