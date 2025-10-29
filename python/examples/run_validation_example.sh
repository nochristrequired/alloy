#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

DEFAULT_LIBRARY_DIR="${PROJECT_ROOT}/dist"
DEFAULT_CONFIGS=(
  "${PROJECT_ROOT}/python/examples/alloy_config_valid.alloy"
  "${PROJECT_ROOT}/python/examples/alloy_config_invalid.alloy"
)

LIBRARY_PATH=""
if (( $# > 0 )); then
  LIBRARY_PATH="$1"
  shift
fi

if [[ -z "${LIBRARY_PATH}" ]]; then
  if [[ -d "${DEFAULT_LIBRARY_DIR}" ]]; then
    mapfile -t CANDIDATES < <(
      find "${DEFAULT_LIBRARY_DIR}" -maxdepth 1 -type f \
        \( -name 'liballoyparser.so' -o -name 'liballoyparser.dylib' -o -name 'liballoyparser.dll' \) |
        sort
    )
  else
    CANDIDATES=()
  fi

  if (( ${#CANDIDATES[@]} == 0 )); then
    echo "error: provide the path to liballoyparser or build it with build_liballoyparser.sh" >&2
    exit 1
  fi

  LIBRARY_PATH="${CANDIDATES[0]}"
fi

if [[ ! -f "${LIBRARY_PATH}" ]]; then
  echo "error: shared library not found at ${LIBRARY_PATH}" >&2
  exit 1
fi

CONFIGS=("$@")
if (( ${#CONFIGS[@]} == 0 )); then
  CONFIGS=("${DEFAULT_CONFIGS[@]}")
fi

for config in "${CONFIGS[@]}"; do
  if [[ ! -f "${config}" ]]; then
    echo "error: configuration file not found at ${config}" >&2
    exit 1
  fi
done

export PYTHONPATH="${PROJECT_ROOT}:${PYTHONPATH:-}"

python "${PROJECT_ROOT}/python/examples/run_validation.py" "${LIBRARY_PATH}" "${CONFIGS[@]}"
