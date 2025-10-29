#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

cd "${PROJECT_ROOT}"

if ! command -v go >/dev/null 2>&1; then
  echo "error: Go toolchain is required but not found in PATH" >&2
  exit 1
fi

if ! command -v make >/dev/null 2>&1; then
  echo "error: make is required but not found in PATH" >&2
  exit 1
fi

make liballoyparser

LIBRARY_PATH=$(find "${PROJECT_ROOT}/dist" -maxdepth 1 -type f \( -name 'liballoyparser.so' -o -name 'liballoyparser.dylib' -o -name 'liballoyparser.dll' \) -print -quit || true)

if [[ -z "${LIBRARY_PATH}" ]]; then
  echo "error: failed to locate liballoyparser artifact in ${PROJECT_ROOT}/dist" >&2
  exit 1
fi

echo "liballoyparser built at ${LIBRARY_PATH}"
