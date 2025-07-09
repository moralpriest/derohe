#!/usr/bin/env bash
set -euo pipefail

# Usage/help
if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  echo "Usage: $0 [tag] [dockerhub-username]"
  echo "Builds all Dero Docker images with optional tag (default: local) and optionally tags/pushes to Docker Hub if username is provided."
  echo "Example: $0 v1.2.3 myuser"
  exit 0
fi

TAG="${1:-local}"
DOCKERHUB_USER="${2:-}"

# Check for Docker
if ! command -v docker &>/dev/null; then
  echo "Error: Docker is not installed or not in PATH." >&2
  exit 1
fi

echo "Building Docker images with tag: $TAG"
if [ -n "$DOCKERHUB_USER" ]; then
  echo "Docker Hub user: $DOCKERHUB_USER (images will be tagged and pushed)"
fi

failures=()

build_image() {
  local dockerfile=$1
  local image=$2
  echo "Building $image:$TAG from $dockerfile"
  if ! docker build -f "$dockerfile" -t "$image:$TAG" .; then
    echo "Failed to build $image:$TAG" >&2
    failures+=("$image")
  fi
  if [ -n "$DOCKERHUB_USER" ]; then
    echo "Tagging $image:$TAG as $DOCKERHUB_USER/$image:$TAG and pushing to Docker Hub"
    docker tag "$image:$TAG" "$DOCKERHUB_USER/$image:$TAG"
    docker push "$DOCKERHUB_USER/$image:$TAG"
  fi
}

build_image cmd/dero-miner/Dockerfile dero-miner
build_image cmd/explorer/Dockerfile dero-explorer
build_image cmd/simulator/Dockerfile dero-simulator
build_image cmd/dero-wallet-cli/Dockerfile dero-wallet-cli
build_image cmd/derod/Dockerfile derod

echo
if [ ${#failures[@]} -eq 0 ]; then
  echo "All Docker images built successfully!"
else
  echo "Some images failed to build: ${failures[*]}" >&2
  exit 1
fi 