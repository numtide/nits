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
        vendorHash = "sha256-juSBLh4Y/1qjg+5ptCWhIKaueqDESNrpI3wXsmNRgeg=";

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
