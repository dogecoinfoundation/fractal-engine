#!/bin/bash

echo "Active Fractal Engine stacks:"
echo "============================="

# List all containers with fractal-stack prefix
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" --filter "name=fractal-stack" | head -n 1
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" --filter "name=fractal-stack" | grep -v NAMES

echo ""
echo "Available commands:"
echo "  ./scripts/run-stack.sh <instance-id>   - Start a new stack"
echo "  ./scripts/stop-stack.sh <instance-id>  - Stop a specific stack"
echo "  ./scripts/stop-all-stacks.sh           - Stop all stacks"
