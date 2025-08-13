{ lib, postgresql, writeShellScriptBin }:

let
  # Create a wrapper script for PostgreSQL
  postgres-wrapper = writeShellScriptBin "fractalstore" ''
    #!/usr/bin/env bash
    
    # Default environment variables
    export POSTGRES_USER=''${POSTGRES_USER:-fractal}
    export POSTGRES_PASSWORD=''${POSTGRES_PASSWORD:-fractal}
    export POSTGRES_DB=''${POSTGRES_DB:-fractal}
    export PGDATA=''${PGDATA:-$HOME/.fractalstore/data}
    
    # Create data directory if it doesn't exist
    mkdir -p "$PGDATA"
    
    # Initialize database if needed
    if [ ! -f "$PGDATA/PG_VERSION" ]; then
      echo "Initializing PostgreSQL database..."
      ${postgresql}/bin/initdb -D "$PGDATA" -U "$POSTGRES_USER" --pwfile=<(echo "$POSTGRES_PASSWORD")
      
      # Start temporary server to create database  
      ${postgresql}/bin/pg_ctl -D "$PGDATA" -l "$PGDATA/server.log" -o "-k /tmp -p ''${PGPORT:-5432}" start
      sleep 3
      PGHOST=/tmp ${postgresql}/bin/createdb -U "$POSTGRES_USER" -p "''${PGPORT:-5432}" "$POSTGRES_DB" || true
      ${postgresql}/bin/pg_ctl -D "$PGDATA" stop
    fi
    
    # Start PostgreSQL
    PORT="''${PGPORT:-5432}"
    echo "Starting PostgreSQL on port $PORT..."
    exec ${postgresql}/bin/postgres -D "$PGDATA" -k /tmp -p "$PORT"
  '';
in

lib.hiPrio postgres-wrapper
