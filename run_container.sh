#!/bin/bash
set -e

# Detect container engine
echo "Checking for Docker or Podman..."
if command -v docker &>/dev/null; then
    ENGINE=docker
elif command -v podman &>/dev/null; then
    ENGINE=podman
else
    echo "Error: Neither Docker nor Podman is installed." >&2
    exit 1
fi

echo "Using $ENGINE."

# Build image
echo "Building image..."
$ENGINE build -t httping .

# Run container
debug_env=""
if [ "$1" = "debug" ]; then
    echo "Enabling DEBUG_PING_OUTPUT=1 in the container."
    debug_env="-e DEBUG_PING_OUTPUT=1"
fi

echo "Running container (exposes :8080)..."
if [ "$ENGINE" = "docker" ]; then
    CAPARG="--cap-add=NET_RAW"
else
    CAPARG="--cap-add=net_raw"
fi
$ENGINE run --rm -p 8080:8080 $CAPARG $debug_env httping
