lib: {
  cleanSourceAndNix = with lib;
    src:
      cleanSourceWith {
        filter = cleanSourceFilter;
        src = cleanSourceWith {
          inherit src;
          filter = name: type: !((type == "directory" && name == "nix") || (hasSuffix ".nix" name));
        };
      };

  mkAgentHost = {
    id,
    self,
    pkgs,
    modules ? [
      ./dev/agents/modules/base.nix
      ./dev/agents/modules/vm.nix
      ./dev/agents/modules/agent.nix
    ],
    extraModules ? [],
  }:
    lib.nixosSystem rec {
      system = pkgs.system;
      inherit pkgs modules;
      specialArgs = {
        inherit self id;
        hostname = "agent-host-${builtins.toString id}";
      };
    };
}
