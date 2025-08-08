#!/bin/bash

echo "Stopping all Fractal Engine stacks..."

# Find all project names with fractal-stack prefix
for container in $(docker ps --format "{{.Names}}" --filter "name=fractal-stack"); do
    # Extract instance ID from container name
    if [[ $container =~ fractal-stack-([0-9]+) ]]; then
        instance_id="${BASH_REMATCH[1]}"
        echo "Stopping stack instance ${instance_id}"
        ./scripts/stop-stack.sh "${instance_id}"
    fi
done

echo "All stacks stopped."
