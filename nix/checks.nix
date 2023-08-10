{
  self,
  lib,
  ...
}: {
  perSystem = {
    self',
    pkgs,
    ...
  }: {
    checks =
      {
        nix-lint =
          pkgs.runCommand "nix-lint" {
            nativeBuildInputs = with pkgs; [deadnix];
          } ''
            cp --no-preserve=mode -r ${self} source
            HOME=$TMPDIR deadnix -f source
            touch $out
          '';
      }
      # merge in the package derivations to force a build of all packages during a `nix flake check`
      // (with lib; mapAttrs' (n: nameValuePair "package-${n}") self'.packages);

    devshells.default = {
      commands = [
        {
          name = "check";
          help = "run all linters and build all packages";
          category = "checks";
          command = "nix flake check";
        }
        {
          name = "fix";
          help = "Remove unused nix code";
          category = "checks";
          command = "${pkgs.deadnix}/bin/deadnix -e $PRJ_ROOT";
        }
      ];
    };
  };
}
