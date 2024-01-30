{inputs, ...}: {
  imports = [
    inputs.treefmt-nix.flakeModule
  ];
  perSystem = {
    config,
    lib,
    pkgs,
    ...
  }: {
    treefmt.config = {
      inherit (config.flake-root) projectRootFile;
      flakeCheck = true;
      flakeFormatter = true;
      programs = {
        gofumpt.enable = true;
        prettier.enable = true;
      };

      settings.formatter = {
        prettier.options = ["--tab-width" "4"];

        # we need to ensure statix and deadnix are run sequentially for a given file
        # currently this is the only way of doing that with treefmt
        nix = {
          command = "${pkgs.bash}/bin/bash";
          options = [
            "-euc"
            ''
              for file in "$@"; do
                  # we rely on defaults for all formatters
                  ${lib.getExe pkgs.deadnix} --edit "$file"
                  ${lib.getExe pkgs.statix} fix "$file"
                  ${lib.getExe pkgs.alejandra} "$file"
              done
            ''
            "--" # bash swallows the second argument when using -c
          ];
          includes = ["*.nix"];
        };
      };
    };

    devshells.default = {
      commands = [
        {
          category = "formatting";
          name = "fmt";
          help = "format the repo";
          command = "nix fmt";
        }
      ];
    };
  };
}
