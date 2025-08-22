{
  description = "Fractal Engine - Configurable Dogecoin services";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        lib = nixpkgs.lib;
      in
      {
        packages = rec {
          # Required services (always included)
          fractalengine = pkgs.callPackage ./nix/fractalengine.nix { };
          fractalstore = pkgs.callPackage ./nix/fractalstore.nix { };

          # Optional services
          dogecoin = pkgs.callPackage ./nix/dogecoin.nix { };
          dogenet = pkgs.callPackage ./nix/dogenet.nix { };
          indexer = pkgs.callPackage ./nix/indexer.nix { };
          indexerstore = pkgs.callPackage ./nix/indexerstore.nix { };

          # Service orchestration

          fractal-stack = pkgs.callPackage ./nix/stack.nix { };

          # Predefined configurations
          minimal = pkgs.buildEnv {
            name = "fractal-minimal";
            paths = [
              fractalengine
              fractalstore
            ];
          };

          full = pkgs.buildEnv {
            name = "fractal-full";
            paths = [
              fractalengine
              fractalstore
              dogecoin
              dogenet
              indexer
              indexerstore
            ];
          };

          # Custom configurable build
          custom =
            {
              withDogecoin ? false,
              withDogenet ? false,
              withIndexer ? false,
            }:
            pkgs.buildEnv {
              name = "fractal-custom";
              paths = [
                fractalengine
                fractalstore
              ]
              ++ lib.optional withDogecoin dogecoin
              ++ lib.optional withDogenet dogenet
              ++ lib.optional withIndexer indexer
              ++ lib.optional withIndexer indexerstore;
            };

          default = minimal;
        };

        # Development shells
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go_1_24
            nodejs_18
            postgresql
            git
            curl
          ];
        };

        # Apps for easy running
        apps = {
          fractal = flake-utils.lib.mkApp {
            drv = self.packages.${system}.fractalengine;
            name = "fractalengine";
          };

          stack = flake-utils.lib.mkApp {
            drv = self.packages.${system}.fractal-stack;
            name = "fractal-stack";
          };
        };
      }
    );
}
