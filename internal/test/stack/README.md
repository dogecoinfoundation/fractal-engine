# E2E Testing Stack

## Quick Start
```sh

# Build stack
nix build .#fractal-stack

# Run stacks
nix run .#stack 1 up
nix run .#stack 2 up

# Check ports/PIDs for stacks
nix run .#stack 1 ports
nix run .#stack 2 ports

# Running tests
go test ./internal/test/stack/...

# Cleanup
nix run .#stack 1 clean
nix run .#stack 2 clean
```
