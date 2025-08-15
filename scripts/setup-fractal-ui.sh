#!/bin/bash

set -e

REPO_URL="https://github.com/dogecoinfoundation/fractal-ui"
TEMP_DIR=$(mktemp -d)

cleanup() {
    echo "Cleaning up temporary directory..."
    rm -rf "$TEMP_DIR"
}

trap cleanup EXIT

echo "Checking out fractal-ui repository to temporary directory..."
echo "Using temp dir: $TEMP_DIR"

git clone "$REPO_URL" "$TEMP_DIR/fractal-ui"

echo "Converting pnpm-lock.yaml to package-lock.json..."

cd "$TEMP_DIR/fractal-ui"

if [ ! -f "pnpm-lock.yaml" ]; then
    echo "Error: pnpm-lock.yaml not found"
    exit 1
fi

if command -v npm >/dev/null 2>&1; then
    echo "Converting pnpm-lock.yaml to package-lock.json..."
    
    # Remove any existing package-lock.json and node_modules
    rm -f package-lock.json
    rm -rf node_modules
    
    # Use npm to install dependencies, which will generate package-lock.json
    echo "Installing dependencies with npm to generate package-lock.json..."
    npm install
    
    if [ -f "package-lock.json" ]; then
        echo "Copying package-lock.json to nix/ folder..."
        cp package-lock.json "$OLDPWD/nix/"
        echo "Conversion completed successfully!"
    else
        echo "Error: package-lock.json was not generated"
        exit 1
    fi
else
    echo "Error: npm not found. Please install npm to convert the lock file."
    exit 1
fi

echo "fractal-ui conversion completed!"
