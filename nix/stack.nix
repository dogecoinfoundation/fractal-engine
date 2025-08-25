{
  pkgs,
  fractalengine,
  fractalstore,
  dogecoin,
  indexer,
  indexerstore,
}:

let
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
