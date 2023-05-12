{
  self,
  lib,
  ...
}: {
  perSystem = {
    self',
    inputs',
    pkgs,
    config,
    ...
  }: {
    checks =
      {
        # TODO statix currently seems to be broken
        #        statix =
        #          pkgs.runCommand "statix" {
        #            nativeBuildInputs = [pkgs.statix];
        #          } ''
        #            cp --no-preserve=mode -r ${self} source
        #            cd source
        #            HOME=$TMPDIR statix check
        #            touch $out
        #          '';
      }
      # merge in the package derivations to force a build of all packages during a `nix flake check`
      // (with lib; mapAttrs' (n: nameValuePair "package-${n}") self'.packages);

    devshells.default = {
      commands = [
        {
          category = "build";
          name = "check";
          help = "run all linters and build all packages";
          command = "nix flake check";
        }
      ];
    };
  };
}
