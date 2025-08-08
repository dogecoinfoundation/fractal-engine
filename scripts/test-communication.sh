#!/bin/bash

# Test communication between stack instances
INSTANCE_A=${1:-1}
INSTANCE_B=${2:-2}

echo "Testing communication between stack ${INSTANCE_A} and stack ${INSTANCE_B}"

# Get the network IP addresses of services
CONTAINER_A="dogecoin-${INSTANCE_A}"
CONTAINER_B="dogecoin-${INSTANCE_B}"

echo "Container A: ${CONTAINER_A}"
echo "Container B: ${CONTAINER_B}"

# Test if containers can see each other on the shared network
echo "Testing network connectivity..."
# Get the IP of container B
IP_B=$(docker network inspect fractal-shared --format '{{range .Containers}}{{if eq .Name "'${CONTAINER_B}'"}}{{.IPv4Address}}{{end}}{{end}}' | cut -d'/' -f1)
echo "Testing connection from ${CONTAINER_A} to ${CONTAINER_B} (${IP_B})"

# Get the actual RPC port from container B's environment
RPC_PORT_B=$(docker exec "${CONTAINER_B}" env | grep RPC_PORT | cut -d'=' -f2)
echo "Container B RPC port: ${RPC_PORT_B}"

# Use nc (netcat) to test TCP connectivity on the RPC port
docker exec "${CONTAINER_A}" sh -c "timeout 3 nc -z ${IP_B} ${RPC_PORT_B}" 2>/dev/null && {
    echo "✓ Stack ${INSTANCE_A} can reach Stack ${INSTANCE_B} on port ${RPC_PORT_B}"
} || {
    echo "✗ Stack ${INSTANCE_A} cannot reach Stack ${INSTANCE_B} on port ${RPC_PORT_B}"
}

# Show the IP addresses in the shared network
echo ""
echo "IP addresses in fractal-shared network:"
docker network inspect fractal-shared --format '{{range .Containers}}{{.Name}}: {{.IPv4Address}}{{"\n"}}{{end}}'
