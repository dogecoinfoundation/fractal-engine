{
  pkgs,
}:

let
  fractalengine = pkgs.callPackage ./fractalengine.nix { };
  fractalstore = pkgs.callPackage ./fractalstore.nix { };
  dogecoin = pkgs.callPackage ./dogecoin.nix { };
  indexer = pkgs.callPackage ./indexer.nix { };
  indexerstore = pkgs.callPackage ./indexerstore.nix { };

  # Stack management script with port isolation
  fractal-stack = pkgs.writeShellScriptBin "fractal-stack" (
    builtins.replaceStrings
      [
        "@postgresql@"
        "@fractalstore@"
        "@indexerstore@"
        "@dogecoin@"
        "@fractalengine@"
        "@indexer@"
      ]
      [
        "${pkgs.postgresql}"
        "${fractalstore}"
        "${indexerstore}"
        "${dogecoin}"
        "${fractalengine}"
        "${indexer}"
      ]
      (builtins.readFile ./stack.sh)
  );

in
fractal-stack
