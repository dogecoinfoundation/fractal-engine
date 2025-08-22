{
  description = "Fractal Engine - Configurable Dogecoin services";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";
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

        dateFormatter = d:
          "${builtins.substring 0 4 d}/${builtins.substring 4 2 d}/${builtins.substring 6 2 d}";

        # Required services (always included)
        fractalengine = pkgs.callPackage ./nix/fractalengine.nix {
          rev = if self ? rev then self.rev else "dirty";
          date = if self ? lastModifiedDate then dateFormatter self.lastModifiedDate else "unknown-date";
        };
        fractalstore = pkgs.callPackage ./nix/fractalstore.nix {};

        # Optional services
        dogecoin = pkgs.callPackage ./nix/dogecoin.nix {};
        dogenet = pkgs.callPackage ./nix/dogenet.nix {};
        indexer = pkgs.callPackage ./nix/indexer.nix {};
        indexerstore = pkgs.callPackage ./nix/indexerstore.nix {};
        fractaladmin = pkgs.callPackage ./nix/fractaladmin.nix {};
      in
      {
        packages = rec {
          inherit fractalengine fractalstore dogecoin dogenet indexer indexerstore fractaladmin;

          # Service orchestration
          fractal-stack = pkgs.callPackage ./nix/stack.nix {
            inherit fractalengine fractalstore dogecoin dogenet indexer indexerstore;
          };

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

          default = fractalengine;
        };

        formatter = pkgs.nixfmt-tree;

        # Development shells
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go_1_24
            nodejs_22
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

          # Development task apps (migrated from Makefile)
          test = {
            type = "app";
            program = "${pkgs.writeShellApplication {
              name = "fractal-test";
              runtimeInputs = [ pkgs.go_1_24 ];
              text = ''
                set -euo pipefail
                IFS=' ' read -r -a EXTRA <<< "''${GO_TEST_EXTRA_FLAGS:-}"
                ENV=test TZ=UTC go test "''${EXTRA[@]}" -p 1 -covermode=count -coverprofile=coverage.txt -timeout=30m ./...
              '';
            }}/bin/fractal-test";
          };

          coverage = {
            type = "app";
            program = "${pkgs.writeShellApplication {
              name = "fractal-coverage";
              runtimeInputs = [ pkgs.go_1_24 ];
              text = ''
                set -euo pipefail
                go tool cover -func=coverage.txt
              '';
            }}/bin/fractal-coverage";
          };

          coverage-html = {
            type = "app";
            program = "${pkgs.writeShellApplication {
              name = "fractal-coverage-html";
              runtimeInputs = [ pkgs.go_1_24 ];
              text = ''
                set -euo pipefail
                go tool cover -html=coverage.txt
              '';
            }}/bin/fractal-coverage-html";
          };

          lint = {
            type = "app";
            program = "${pkgs.writeShellApplication {
              name = "fractal-lint";
              runtimeInputs = [ pkgs.golangci-lint ];
              text = ''
                set -euo pipefail
                golangci-lint run --fix
              '';
            }}/bin/fractal-lint";
          };

          format = {
            type = "app";
            program = "${pkgs.writeShellApplication {
              name = "fractal-format";
              runtimeInputs = [ pkgs.golangci-lint ];
              text = ''
                set -euo pipefail
                golangci-lint fmt
              '';
            }}/bin/fractal-format";
          };

          tidy = {
            type = "app";
            program = "${pkgs.writeShellApplication {
              name = "fractal-tidy";
              runtimeInputs = [ pkgs.go_1_24 ];
              text = ''
                set -euo pipefail
                go mod tidy
              '';
            }}/bin/fractal-tidy";
          };
        };
      }
    );
}
