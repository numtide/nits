final: _: {
  cleanSourceAndNix = with final;
    src:
      cleanSourceWith {
        filter = cleanSourceFilter;
        src = cleanSourceWith {
          inherit src;
          filter = name: type: !((type == "directory" && name == "nix") || (hasSuffix ".nix" name));
        };
      };
}
