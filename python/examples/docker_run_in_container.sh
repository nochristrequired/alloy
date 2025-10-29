#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

cd "${PROJECT_ROOT}"

"${SCRIPT_DIR}/build_liballoyparser.sh"

"${SCRIPT_DIR}/run_validation_example.sh" "$@"
