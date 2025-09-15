#!/usr/bin/env bash

set -euo pipefail

# Usage: fractal-stack <instance-id> [up|down|ps|logs]
INSTANCE_ID=${1:-1}
COMMAND=${2:-up}

shift 2 2>/dev/null || shift $# # Remove first two args, pass rest to indexer
INDEXER_ARGS="$*"

# Validate instance ID
if ! [[ "$INSTANCE_ID" =~ ^[0-9]+$ ]]; then
  echo "Error: Instance ID must be a number"
  exit 1
fi

PROJECT_NAME="fractal-stack-$INSTANCE_ID"

# Port ranges - each instance gets a block of 100 ports
BASE_PORT=$((8600 + (INSTANCE_ID * 100)))
DOGE_RPC_PORT=$((8700 + 14556))
DOGE_ZMQ_PORT=$((BASE_PORT + 20000))
FRACTAL_ENGINE_PORT=$((BASE_PORT + 20))
DOGENET_PORT=$((BASE_PORT + 30))
DOGENET_WEB_PORT=$((BASE_PORT + 40))
INDEXER_PORT=$((BASE_PORT + 50))
POSTGRES_PORT=$((BASE_PORT + 60))
DOGENET_HANDLER_PORT=$((BASE_PORT + 70))
DOGENET_BIND_PORT=$((BASE_PORT + 73))
DOGE_P2P_PORT=$((BASE_PORT + 80))

# Data directories for instance isolation
BASE_DIR="$HOME/.fractal-stack-$INSTANCE_ID"
DOGECOIN_DATA="$BASE_DIR/dogecoin-$INSTANCE_ID"
POSTGRES_DATA="$BASE_DIR/postgres-$INSTANCE_ID"
INDEXER_DATA="$BASE_DIR/indexer-$INSTANCE_ID"
DOGENET_DATA="$BASE_DIR/dogenet-$INSTANCE_ID"
LOGS_DIR="$BASE_DIR/logs"
PIDS_FILE="$BASE_DIR/pids"

INDEXER_DB_URL=$BASE_DIR/indexerstore/indexer.db

# Create instance directories
mkdir -p "$BASE_DIR" "$POSTGRES_DATA" "$DOGECOIN_DATA" "$INDEXER_DATA" "$DOGENET_DATA" "$LOGS_DIR"
mkdir -p $BASE_DIR/indexerstore

export POSTGRES_USER=fractalstore
export POSTGRES_PASSWORD=fractalstore
export POSTGRES_DB=fractalstore
export PGDATA="$POSTGRES_DATA"
export PGPORT=$POSTGRES_PORT

export FRACTAL_ENGINE_HOST=localhost
export FRACTAL_ENGINE_DB="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:$POSTGRES_PORT/${POSTGRES_DB}"

export DOGECOIN_RPC_HOST=0.0.0.0
export DOGECOIN_RPC_PORT=$DOGE_RPC_PORT
export DOGECOIN_RPC_USER=dogecoinrpc
export DOGECOIN_RPC_PASSWORD=changeme1
export DOGECOIN_DATADIR="$DOGECOIN_DATA"

export DOGE_NET_HANDLER="0.0.0.0:$DOGENET_HANDLER_PORT"
export DOGENET_WEB_PORT=$DOGENET_WEB_PORT
export DOGENET_BIND_HOST=0.0.0.0
export DOGENET_BIND_PORT=$DOGENET_BIND_PORT
export DOGENET_DB_FILE="$BASE_DIR/dogenet.db"

export INDEXER_DOGECOIN_RPC="http://dogecoinrpc:changeme1@0.0.0.0:$DOGE_RPC_PORT"
export INDEXER_PORT=$INDEXER_PORT
export INDEXER_STARTINGHEIGHT=0

cleanup() {
  echo "Stopping stack instance $INSTANCE_ID..."
  if [ -f "$PIDS_FILE" ]; then
    while read -r service pid; do
      if kill -0 "$pid" 2>/dev/null; then
        echo "Stopping $service (PID: $pid)"
        kill "$pid" || true
      fi
    done < "$PIDS_FILE"
    rm -f "$PIDS_FILE"
  fi

    if [ -f "$POSTGRES_DATA/postmaster.pid" ]; then
        @postgresql@/bin/pg_ctl -D "$POSTGRES_DATA" stop || true
    fi
}


start_service() {
  local service_name="$1"
  local cmd="$2"
  local log_file="$LOGS_DIR/${service_name}.log"
  {
    echo "===== [$PROJECT_NAME] $service_name starting at $(date -Is) ====="
    echo "Command: $cmd"
  } >> "$log_file"

  echo "[$PROJECT_NAME] Starting $service_name on various ports... (logging to $log_file)"

  # Start the command with stdout/stderr redirected to the per-service log file
  $cmd >> "$log_file" 2>&1 &

  local pid=$!
  echo "$service_name $pid" >> "$PIDS_FILE"
  echo "[$PROJECT_NAME] $service_name started with PID $pid"
}


show_status() {
  echo "=== Fractal Stack Instance $INSTANCE_ID ==="
  echo "Project: $PROJECT_NAME"
  echo "Data directory: $BASE_DIR"
  echo "Port assignments:"
  echo "  Dogecoin RPC:     $DOGE_RPC_PORT"
  echo "  Dogecoin ZMQ:     $DOGE_ZMQ_PORT"
  echo "  Fractal Engine:   $FRACTAL_ENGINE_PORT"
  echo "  PostgreSQL:       $POSTGRES_PORT"
  echo "  Dogenet:          $DOGENET_PORT"
  echo "  Dogecoin P2P:     $DOGE_P2P_PORT"
  echo "  Dogenet Web:      $DOGENET_WEB_PORT"
  echo "  Dogenet Bind:     $DOGENET_BIND_PORT"
  echo "  Indexer API:      $INDEXER_PORT"
  echo ""

  if [ -f "$PIDS_FILE" ]; then
    echo "Running services:"
    while read -r service pid; do
      if kill -0 "$pid" 2>/dev/null; then
        echo "  ✓ $service (PID: $pid)"
      else
        echo "  ✗ $service (dead)"
      fi
    done < "$PIDS_FILE"
  else
    echo "No services running"
  fi
}

case "$COMMAND" in
  up)
    trap cleanup EXIT INT TERM

    echo "Starting Fractal Stack Instance $INSTANCE_ID..."
    show_status

    # Start PostgreSQL instances first
    start_service "fractalstore" "env POSTGRES_USER=$POSTGRES_USER POSTGRES_PASSWORD=$POSTGRES_PASSWORD POSTGRES_DB=$POSTGRES_DB PGDATA=$PGDATA PGPORT=$POSTGRES_PORT @fractalstore@/bin/fractalstore"

    # Wait for fractalstore PostgreSQL to be ready
    echo "Waiting for fractalstore PostgreSQL to be ready on port $POSTGRES_PORT..."
    timeout=30
    while [ $timeout -gt 0 ]; do
      if @postgresql@/bin/pg_isready -h localhost -p $POSTGRES_PORT >/dev/null 2>&1; then
        echo "Fractalstore PostgreSQL is ready!"
        break
      fi
      echo "Waiting for fractalstore PostgreSQL... ($timeout seconds left)"
      sleep 2
      timeout=$((timeout - 2))
    done

    if [ $timeout -le 0 ]; then
      echo "ERROR: Fractalstore PostgreSQL failed to start within 30 seconds"
      exit 1
    fi

    # Start Dogecoin
    if [ "$INSTANCE_ID" = "1" ]; then
      start_service "dogecoin" "env ZMQ_PORT=$DOGE_ZMQ_PORT P2P_PORT=$DOGE_P2P_PORT RPC_USER=$DOGECOIN_RPC_USER RPC_PASSWORD=$DOGECOIN_RPC_PASSWORD RPC_PORT=$DOGECOIN_RPC_PORT INSTANCE_ID=$INSTANCE_ID @dogecoin@/bin/dogecoind"

        # Wait for Dogecoin RPC to be ready
        echo "Waiting for Dogecoin RPC to be ready on port $DOGE_RPC_PORT..."
        timeout=60
        while [ $timeout -gt 0 ]; do
          if curl -s --user "$DOGECOIN_RPC_USER:$DOGECOIN_RPC_PASSWORD" \
             --data-binary '{"jsonrpc":"1.0","id":"test","method":"getblockchaininfo","params":[]}' \
             -H 'content-type: text/plain;' \
             "http://localhost:$DOGE_RPC_PORT/" >/dev/null 2>&1; then
            echo "Dogecoin RPC is ready!"
            break
          fi
          echo "Waiting for Dogecoin RPC... ($timeout seconds left)"
          sleep 2
          timeout=$((timeout - 2))
        done

        if [ $timeout -le 0 ]; then
          echo "ERROR: Dogecoin RPC failed to start within 60 seconds"
          exit 1
        fi
    fi

    echo " --rpc-server-host 0.0.0.0 \
    --rpc-server-port $FRACTAL_ENGINE_PORT \
    --doge-net-network unix \
    --doge-net-address $BASE_DIR/dogenet.sock \
    --doge-net-web-address 0.0.0.0:$DOGENET_WEB_PORT \
    --doge-scheme http \
    --doge-host localhost \
    --doge-port $DOGE_RPC_PORT \
    --doge-user $DOGECOIN_RPC_USER \
    --doge-password $DOGECOIN_RPC_PASSWORD \
    --database-url $FRACTAL_ENGINE_DB?sslmode=disable \
    --embed-dogenet true"

    start_service "fractalengine" "@fractalengine@/bin/fractal-engine \
      --rpc-server-host 0.0.0.0 \
      --rpc-server-port $FRACTAL_ENGINE_PORT \
      --doge-net-network unix \
      --doge-net-address $BASE_DIR/dogenet.sock \
      --doge-net-web-address 0.0.0.0:$DOGENET_WEB_PORT \
      --doge-net-db-file $DOGENET_DB_FILE \
      --doge-scheme http \
      --doge-host localhost \
      --doge-port $DOGE_RPC_PORT \
      --doge-user $DOGECOIN_RPC_USER \
      --doge-password $DOGECOIN_RPC_PASSWORD \
      --database-url $FRACTAL_ENGINE_DB?sslmode=disable \
      --embed-dogenet true"

    rm -rf $INDEXER_DB_URL

    if [ "$INSTANCE_ID" = "1" ]; then
    start_service "indexer" "@indexer@/bin/indexer \
      -bindapi localhost:${INDEXER_PORT} \
      -dburl $INDEXER_DB_URL \
      -chain regtest \
      -rpchost localhost \
      -rpcpass $DOGECOIN_RPC_PASSWORD \
      -rpcport $DOGE_RPC_PORT \
      -rpcuser $DOGECOIN_RPC_USER \
      -zmqhost localhost \
      -zmqport $DOGE_ZMQ_PORT \
      -startingheight $INDEXER_STARTINGHEIGHT \
      $INDEXER_ARGS"

    fi

    echo ""
    echo "=== Stack $INSTANCE_ID Ready ==="
    echo "Fractal Engine: http://localhost:$FRACTAL_ENGINE_PORT"
    echo "Dogenet Web:    http://localhost:$DOGENET_WEB_PORT"
    echo "Indexer API:    http://localhost:$INDEXER_PORT"
    echo ""
    echo "Press Ctrl+C to stop all services"
    wait
    ;;


  down)

    cleanup

    echo "Purging PostgreSQL data directories for instance $INSTANCE_ID..."
    rm -rf "$POSTGRES_DATA"

    ;;


  ps|status)
    show_status
    ;;


  ports)

    echo "=== Port Usage for Stack Instance $INSTANCE_ID ==="

    if [ -f "$PIDS_FILE" ]; then

      echo "Service → PID → Listening On"

      echo "──────────────────────────────"


      while read -r service pid; do
        pid=$(printf '%s' "$pid" | tr -cd '0-9')
        if [ -z "$pid" ]; then
          continue
        fi


        if kill -0 "$pid" 2>/dev/null; then

          if command -v ss >/dev/null 2>&1; then
            addresses=$(
              { ss -ltnp 2>/dev/null; ss -lunp 2>/dev/null; } \
              | awk -v pid="$pid" '$0 ~ ("pid=" pid "[,)]") { print $4 }' \
              | sort -u | paste -sd, - 2>/dev/null
            )
          elif command -v netstat >/dev/null 2>&1; then
            addresses=$(
              { netstat -ltnp 2>/dev/null; netstat -lunp 2>/dev/null; } \
              | awk -v pid="$pid" '$7 ~ ("^" pid "/") { print $4 }' \
              | sort -u | paste -sd, - 2>/dev/null
            )
          else
            addresses=""
          fi
          if [ -n "$addresses" ]; then

            printf "%-12s → %-6s → %s\n" "$service" "$pid" "$addresses"

          else

            printf "%-12s → %-6s → (no listening ports)\n" "$service" "$pid"

          fi

        else

          printf "%-12s → %-6s → (dead process)\n" "$service" "$pid"

        fi

      done < "$PIDS_FILE"

    else

      echo "No services running (PID file not found)"

    fi

    ;;


  logs)
    echo "Service logs for instance $INSTANCE_ID:"
    echo "Data directory: $BASE_DIR"
    echo "Check individual service outputs or $BASE_DIR/logs/ if implemented"
    ;;

  clean)
    cleanup
    echo "Removing data directory: $BASE_DIR"
    rm -rf "$BASE_DIR"
    rm -rf "$DOGENET_DATA"
    rm -rf "$PGDATA"
    rm -rf "/tmp/dogecoin$INSTANCE_ID"
    ;;

  *)
    echo "Usage: $0 <instance-id> [up|down|ps|ports|logs|clean] [indexer-args...]"
    echo ""
    echo "Commands:"
    echo "  up     - Start the stack"
    echo "  down   - Stop the stack"
    echo "  ps     - Show status"
    echo "  ports  - Show network ports used by each service"
    echo "  logs   - Show log locations"
    echo "  clean  - Stop and remove all data"
    echo ""
    echo "Examples:"
    echo "  $0 1 up                    # Start instance 1"
    echo "  $0 2 up                    # Start instance 2 (different ports)"
    echo "  $0 1 ports                 # Show port usage for instance 1"
    echo "  $0 1 up --verbose --debug  # Start with indexer args"
    echo "  $0 1 down                  # Stop instance 1"
    exit 1
    ;;
esac
