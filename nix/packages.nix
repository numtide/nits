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

        src = ../.;
        vendorSha256 = "sha256-vZJNXK+bTUqNhqbiIeuwHG1hfLz69DCRFDHLHZ3OVgg=";

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
