#!/bin/bash

echo "Stopping all Fractal Engine stacks..."

# Find all project names with fractal-stack prefix
for container in $(docker-compose ls | grep fractal | awk '{print $1}'); do
    docker-compose -p $container down
done

echo "All stacks stopped."
