#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

if ! command -v docker >/dev/null 2>&1; then
  echo "error: docker is required but not found in PATH" >&2
  exit 1
fi

if [[ $# -gt 0 ]]; then
  IMAGE_TAG="$1"
  shift
else
  IMAGE_TAG="alloy-parser-example:latest"
fi

MOUNTED_PROJECT="${PROJECT_ROOT}"

DOCKER_CMD=(
  docker run --rm \
    -v "${MOUNTED_PROJECT}:/workspace" \
    -w /workspace \
    "${IMAGE_TAG}" \
    "/workspace/python/examples/docker_run_in_container.sh"
)

if [[ $# -gt 0 ]]; then
  DOCKER_CMD+=("$@")
fi

"${DOCKER_CMD[@]}"
