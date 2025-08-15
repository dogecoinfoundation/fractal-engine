{ lib, postgresql, writeShellScriptBin }:

let
  # Create a wrapper script for PostgreSQL
  postgres-wrapper = writeShellScriptBin "fractalstore" ''
    #!/usr/bin/env bash

    # Create data directory if it doesn't exist
    mkdir -p "$PGDATA"

    # Initialize database if needed
    PORT="''${PGPORT:-5432}"
    echo "DEBUG: All env vars:"
    echo "  POSTGRES_USER=$POSTGRES_USER"
    echo "  POSTGRES_PASSWORD=$POSTGRES_PASSWORD"
    echo "  POSTGRES_DB=$POSTGRES_DB"
    echo "  PGDATA=$PGDATA"
    echo "  PGPORT=$PGPORT"
    echo "  PORT=$PORT"
    if [ ! -f "$PGDATA/PG_VERSION" ]; then
      echo "Initializing PostgreSQL database..."
      ${postgresql}/bin/initdb -D "$PGDATA" -U "$POSTGRES_USER" --pwfile=<(echo "$POSTGRES_PASSWORD")

      # Start temporary server to create database
      echo "Starting temporary PostgreSQL server..."
      ${postgresql}/bin/pg_ctl -D "$PGDATA" -l "$PGDATA/server.log" -o "-p $PORT -k /tmp" start
      sleep 3

      echo "Creating database '$POSTGRES_DB'..."
      ${postgresql}/bin/createdb -h localhost -p $PORT -U "$POSTGRES_USER" "$POSTGRES_DB"
      if [ $? -eq 0 ]; then
        echo "Database '$POSTGRES_DB' created successfully"
      else
        echo "Failed to create database '$POSTGRES_DB'"
        cat "$PGDATA/server.log"
      fi

      ${postgresql}/bin/pg_ctl -D "$PGDATA" stop
      sleep 1
    else
      echo "Database already initialized"
    fi

    # Start PostgreSQL
    echo "Starting PostgreSQL on port $PORT..."
    exec ${postgresql}/bin/postgres -D "$PGDATA" -k /tmp -p "$PORT"
  '';
in

lib.hiPrio postgres-wrapper
