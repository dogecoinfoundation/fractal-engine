#!/bin/bash

# Usage: ./stop-stack.sh <instance-id>
INSTANCE_ID=${1:-1}
PROJECT_NAME="fractal-stack-${INSTANCE_ID}"

echo "Stopping stack instance ${INSTANCE_ID}"

# Stop and remove containers for this instance
docker-compose -p ${PROJECT_NAME} --profile deps --profile fractal down

echo "Stack ${INSTANCE_ID} stopped."
