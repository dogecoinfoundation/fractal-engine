# Nix Configuration for Fractal Engine

This repository provides configurable Nix builds for the Fractal Engine services, allowing you to build only what you need.

## Quick Start

```bash
# Build minimal (required services only: fractalengine + fractalstore)
nix build

# Build everything
nix build .#full

# Build specific services
nix build .#dogecoin
nix build .#fractaladmin
```

## Available Packages

### Required Services (Always Included)
- **fractalengine** - Main Go service
- **fractalstore** - PostgreSQL database wrapper

### Optional Services
- **dogecoin** - Dogecoin node (v1.14.9)
- **dogenet** - Networking service
- **balance-master** - Balance tracking service  
- **indexer** - Blockchain indexer
- **fractaladmin** - Web UI interface

### Predefined Configurations

- `.#minimal` (default) - fractalengine + fractalstore
- `.#full` - All services

### Custom Builds

```bash
# Custom build with specific services
nix build --expr '
  (import ./flake.nix).outputs.packages.x86_64-linux.custom {
    withDogecoin = true;
    withAdmin = true;
  }
'
```

## Development

```bash
# Enter development shell with all dependencies
nix develop

# Run individual services
nix run .#fractalengine
```

## Service Wrappers

Each service includes a startup wrapper script:
- `fractalstore` - Sets up and runs PostgreSQL
- `dogecoin-regtest` - Runs Dogecoin in regtest mode
- `dogenet-start` - Starts Dogenet with key generation
- `balance-master-start` - Starts balance master service
- `indexer-start` - Starts indexer service
- `fractaladmin` - Starts web interface

## Environment Variables

Services respect the same environment variables as their Docker counterparts. See individual service files for details.

## TODO

- [ ] Add correct SHA256 hashes for external dependencies
- [ ] Test all service builds
- [ ] Add NixOS service modules
- [ ] Add container images via nix2container
