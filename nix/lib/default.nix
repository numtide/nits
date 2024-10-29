{ inputs, ... }:
let
  inherit (builtins) baseNameOf toString elemAt;
  inherit (inputs.nixpkgs.lib) toInt;
in {
  mkTestVM = path:
    let
      folderName = baseNameOf (toString path);

      # "x86_64-linux_test-vm-3-no-hello" -> [ "x86_64-linux" "3" "-no-hello" ]
      # "x86_64-linux_test-vm-3-no-hello" -> [ "x86_64-linux" "3" "" ]
      params = builtins.match "([A-Za-z0-9_\-]+)_test-vm-+([0-9]+)(.*)" folderName;

      platform = elemAt params 0;
      id = toInt(elemAt params 1);
      hello = elemAt params 2;
    in {
      _module.args = {
        inherit id;
        hostname = "test-vm-${toString id}";
      };

      imports = [
        ../dev/modules/base.nix
        ../dev/modules/agent.nix
      ];

      nixpkgs.hostPlatform = platform;

      systemd.services.hello.enable = hello != "-no-hello";
    };
}
