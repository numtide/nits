{
  perSystem = {
    lib,
    pkgs,
    ...
  }: {
    packages = rec {
      nits = pkgs.buildGoModule rec {
        pname = "nits";
        version = "0.0.1+dev";

        src = ../.;
        vendorSha256 = "sha256-ELJtZ2gLZY2aOIx0o0Uh4yhiZfpIAvTtXrFnmiqclec=";

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
  };
}
