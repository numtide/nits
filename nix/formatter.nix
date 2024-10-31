{
  pkgs,
  inputs,
  ...
}:
inputs.treefmt-nix.lib.mkWrapper pkgs {
  projectRootFile = ".git/config";

  programs = {
    alejandra.enable = true;
    deadnix.enable = true;
    gofumpt.enable = true;
    prettier.enable = true;
    statix.enable = true;
  };

  settings = {
    global.excludes = [
      "LICENSE"
      # unsupported extensions
      "*.{gif,png,svg,tape,mts,lock,mod,sum,toml,env,envrc,gitignore,pub,sec}"
    ];

    formatter = {
      deadnix = {
        priority = 1;
      };

      statix = {
        priority = 2;
      };

      alejandra = {
        priority = 3;
      };

      prettier = {
        options = [
          "--tab-width"
          "4"
        ];
      };
    };
  };
}
