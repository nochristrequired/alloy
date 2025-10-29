#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

if ! command -v docker >/dev/null 2>&1; then
  echo "error: docker is required but not found in PATH" >&2
  exit 1
fi

IMAGE_TAG="${1:-alloy-parser-example:latest}"

DOCKERFILE_PATH="${SCRIPT_DIR}/Dockerfile"

docker build -f "${DOCKERFILE_PATH}" -t "${IMAGE_TAG}" "${PROJECT_ROOT}"

echo "Docker image built: ${IMAGE_TAG}"
