{ lib, writeShellScriptBin, pkgs }:

let
  fractalengine = pkgs.callPackage ./fractalengine.nix {};
  fractalstore = pkgs.callPackage ./fractalstore.nix {};
  dogecoin = pkgs.callPackage ./dogecoin.nix {};
  dogenet = pkgs.callPackage ./dogenet.nix {};
  indexer = pkgs.callPackage ./indexer.nix {};
  fractaladmin = pkgs.callPackage ./fractaladmin.nix {};

  # Configuration file for all services
  config = pkgs.writeText "fractal-compose.conf" ''
    # Fractal Engine Configuration
    export FRACTAL_ENGINE_HOST=localhost
    export FRACTAL_ENGINE_PORT=8080
    export FRACTAL_ENGINE_DB=postgresql://fractal:fractal@localhost:5432/fractal
    
    # Fractal Store Configuration  
    export POSTGRES_USER=fractal
    export POSTGRES_PASSWORD=fractal
    export POSTGRES_DB=fractal
    export PGDATA=$HOME/.fractalstore/data
    
    # Dogecoin Configuration
    export DOGECOIN_RPC_HOST=localhost
    export DOGECOIN_RPC_PORT=22555
    export DOGECOIN_RPC_USER=dogecoinrpc
    export DOGECOIN_RPC_PASSWORD=changeme
    
    # Dogenet Configuration
    export DOGE_NET_HANDLER=unix:///tmp/dogenet.sock
    export DOGENET_WEB_PORT=8085
    export DOGENET_BIND_HOST=0.0.0.0
    export DOGENET_BIND_PORT=42000
    
    # Indexer Configuration
    export INDEXER_DOGECOIN_RPC=http://dogecoinrpc:changeme@localhost:22555
    export INDEXER_ENGINE_URL=http://localhost:8080
    
    # Fractal Admin Configuration
    export DATABASE_URL="file:$HOME/.fractaladmin/dev.db"
    export NEXT_TELEMETRY_DISABLED=1
    export PORT=3000
  '';

  # Service management script
  fractal-compose = writeShellScriptBin "fractal-compose" ''
    #!/usr/bin/env bash
    
    set -euo pipefail
    
    # Load configuration
    source ${config}
    
    # Create necessary directories
    mkdir -p $HOME/.fractalstore/data
    mkdir -p $HOME/.fractaladmin
    mkdir -p $HOME/.dogecoin
    mkdir -p $HOME/.indexer
    mkdir -p /tmp
    
    # PID tracking
    PIDS_FILE="$HOME/.fractal-compose.pids"
    
    cleanup() {
      echo "Stopping all services..."
      if [ -f "$PIDS_FILE" ]; then
        while read -r pid; do
          if kill -0 "$pid" 2>/dev/null; then
            kill "$pid" || true
          fi
        done < "$PIDS_FILE"
        rm -f "$PIDS_FILE"
      fi
    }
    
    trap cleanup EXIT INT TERM
    
    start_service() {
      local name="$1"
      local cmd="$2"
      echo "Starting $name..."
      $cmd &
      local pid=$!
      echo "$pid" >> "$PIDS_FILE"
      echo "$name started with PID $pid"
    }
    
    case "''${1:-up}" in
      up)
        echo "Starting Fractal Engine stack..."
        
        # Start PostgreSQL first
        start_service "fractalstore" "${fractalstore}/bin/fractalstore"
        sleep 3
        
        # Start core services
        start_service "fractalengine" "${fractalengine}/bin/fractal-engine"
        start_service "dogecoin" "${dogecoin}/bin/dogecoind -regtest"
        sleep 5
        
        # Start optional services
        start_service "dogenet" "${dogenet}/bin/dogenet-start"
        start_service "indexer" "${indexer}/bin/indexer"
        start_service "fractaladmin" "${fractaladmin}/bin/fractaladmin"
        
        echo "All services started. Press Ctrl+C to stop."
        wait
        ;;
        
      down)
        echo "Stopping services..."
        cleanup
        ;;
        
      logs)
        echo "Service logs are written to individual service outputs"
        ;;
        
      ps)
        echo "Running services:"
        if [ -f "$PIDS_FILE" ]; then
          while read -r pid; do
            if kill -0 "$pid" 2>/dev/null; then
              ps -p "$pid" -o pid,cmd --no-headers
            fi
          done < "$PIDS_FILE"
        else
          echo "No services running"
        fi
        ;;
        
      *)
        echo "Usage: $0 {up|down|logs|ps}"
        exit 1
        ;;
    esac
  '';

in fractal-compose
