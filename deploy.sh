#!/bin/bash

IMAGE_NAME="pingpulse"
REGISTRY="registry.local.timmybtech.com"
TAG="latest"
PLATFORM="linux/amd64"
FULL_IMAGE_NAME="$REGISTRY/$IMAGE_NAME:$TAG"

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

echo "Building image for platform $PLATFORM..."
$ENGINE build --platform $PLATFORM -t $IMAGE_NAME . --no-cache

echo "Tagging image..."
$ENGINE tag $IMAGE_NAME $FULL_IMAGE_NAME

echo "Pushing image to registry..."
$ENGINE push $FULL_IMAGE_NAME

echo "Deployment script completed successfully!"
