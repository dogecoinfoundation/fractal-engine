#!/bin/bash

# Usage: ./run-stack.sh <instance-id> [profiles]
INSTANCE_ID=${1:-1}
PROFILES=${2:-"deps,fractal"}
PROJECT_NAME="fractal-stack-${INSTANCE_ID}"

# Port ranges - each instance gets a block of 100 ports
BASE_PORT=$((8000 + (INSTANCE_ID * 100)))
DOGE_PORT=$((BASE_PORT + 14556))
FRACTAL_PORT=$((BASE_PORT + 2))
DOGENET_PORT=$((BASE_PORT + 3))
DOGENET_WEB_PORT=$((BASE_PORT + 4))
BALANCE_MASTER_PORT=$((BASE_PORT + 5))
POSTGRES_PORT=$((BASE_PORT + 6))
DOGENET_HANDLER_PORT=$((BASE_PORT + 7))
# Subnet ranges - each instance gets a unique subnet
SUBNET_BASE=$((100 + INSTANCE_ID))
SUBNET="192.168.${SUBNET_BASE}.0/24"

DOGENET_HOST="${INSTANCE_ID}-dogenet"
DOGENET_IP="192.168.${SUBNET_BASE}.10"

# Create shared network for inter-stack communication (if it doesn't exist)
docker network create fractal-shared 2>/dev/null || true

# Prepare Go module cache if needed
GO_CACHE_DIR=$(go env GOMODCACHE)
if [ ! -d "$GO_CACHE_DIR" ] || [ -z "$(ls -A "$GO_CACHE_DIR" 2>/dev/null)" ]; then
    echo "Go module cache is empty. Preparing cache..."
    ./scripts/prepare-go-cache.sh
fi

# Use cache-enabled Dockerfiles
BALANCE_MASTER_DOCKERFILE="Dockerfile.balance-master"

# Set build args to handle network issues
# export DOCKER_BUILDKIT=1
export BUILDKIT_PROGRESS=plain
export COMPOSE_DOCKER_CLI_BUILD=1

echo "Starting stack instance ${INSTANCE_ID} with profiles: ${PROFILES}"
echo "Ports: Doge=${DOGE_PORT}, Fractal=${FRACTAL_PORT}, DogeNet=${DOGENET_PORT}/${DOGENET_WEB_PORT}, BalanceMaster=${BALANCE_MASTER_PORT}, Postgres=${POSTGRES_PORT}"
echo "Subnet: ${SUBNET}"
echo "Using Go module cache: $(go env GOMODCACHE)"

# Run docker-compose with custom project name, profiles, and ports
DOGE_NET_NETWORK=tcp \
DOGE_NET_ADDRESS=${DOGENET_IP}:${DOGENET_HANDLER_PORT} \
DOGE_NET_HANDLER="${DOGENET_IP}:${DOGENET_HANDLER_PORT}" \
BALANCE_MASTER_DOCKERFILE=${BALANCE_MASTER_DOCKERFILE} \
DOGE_PORT=${DOGE_PORT} \
FRACTAL_PORT=${FRACTAL_PORT} \
DOGENET_PORT=${DOGENET_PORT} \
DOGENET_WEB_PORT=${DOGENET_WEB_PORT} \
BALANCE_MASTER_PORT=${BALANCE_MASTER_PORT} \
POSTGRES_PORT=${POSTGRES_PORT} \
SUBNET=${SUBNET} \
SUBNET_BASE=${SUBNET_BASE} \
INSTANCE_ID=${INSTANCE_ID} \
docker-compose -p ${PROJECT_NAME} --profile deps --profile fractal up -d --build

echo "Stack ${INSTANCE_ID} started. Containers:"
docker-compose -p ${PROJECT_NAME} ps
