{ lib, writeShellScriptBin, pkgs }:

let
  fractalengine = pkgs.callPackage ./fractalengine.nix {};
  fractalstore = pkgs.callPackage ./fractalstore.nix {};
  dogecoin = pkgs.callPackage ./dogecoin.nix {};
  dogenet = pkgs.callPackage ./dogenet.nix {};
  indexer = pkgs.callPackage ./indexer.nix {};
  indexerstore = pkgs.callPackage ./indexerstore.nix {};
  fractaladmin = pkgs.callPackage ./fractaladmin.nix {};

  # Stack management script with port isolation  
  fractal-stack = pkgs.writeShellScriptBin "fractal-stack" (
    builtins.replaceStrings 
      ["@postgresql@" "@fractalstore@" "@indexerstore@" "@dogecoin@" "@dogenet@" "@fractalengine@" "@indexer@" "@fractaladmin@"]
      ["${pkgs.postgresql}" "${fractalstore}" "${indexerstore}" "${dogecoin}" "${dogenet}" "${fractalengine}" "${indexer}" "${fractaladmin}"]
      (builtins.readFile ./stack.sh)
  );

in fractal-stack
