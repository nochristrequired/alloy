"""Utility to validate Alloy configuration files via the shared parser."""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path
from typing import Iterable

PROJECT_ROOT = Path(__file__).resolve().parents[2]
if str(PROJECT_ROOT) not in sys.path:
    sys.path.insert(0, str(PROJECT_ROOT))

from python.alloy_ffi import AlloyParserBridge, ParseStatus


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "library",
        type=Path,
        help="Path to the liballoyparser shared library (liballoyparser.so/.dylib/.dll).",
    )
    parser.add_argument(
        "configs",
        nargs="+",
        type=Path,
        help="One or more Alloy configuration files to validate.",
    )
    parser.add_argument(
        "--show-payload",
        action="store_true",
        help="Pretty print the full JSON payload returned by the parser.",
    )
    return parser.parse_args()


def validate_configs(library: Path, configs: Iterable[Path], show_payload: bool) -> int:
    try:
        bridge = AlloyParserBridge(library)
    except FileNotFoundError:
        print(f"error: shared library not found at {library}")
        return 1

    exit_code = 0
    for config_path in configs:
        data = config_path.read_bytes()
        result = bridge.parse_file(config_path.name, data)

        status = result.status
        diagnostics = result.payload.get("diagnostics", [])

        print(f"\n=== {config_path} ===")
        print(f"status: {status.name}")
        if diagnostics:
            print("diagnostics:")
            for entry in diagnostics:
                severity = entry.get("severity", "unknown")
                message = entry.get("message", "")
                start = entry.get("start", {})
                line = start.get("line")
                column = start.get("column")
                location = f" (line {line}, column {column})" if line is not None else ""
                print(f"  - {severity}{location}: {message}")
        else:
            print("diagnostics: none")

        if show_payload:
            print("payload:")
            print(json.dumps(result.payload, indent=2))

        if status != ParseStatus.OK:
            exit_code = 2

    return exit_code


def main() -> int:
    args = parse_args()
    configs = [path for path in args.configs if path.exists()]
    missing = [path for path in args.configs if not path.exists()]

    for path in missing:
        print(f"warning: skipping missing configuration {path}")

    if not configs:
        print("error: no configuration files to validate")
        return 1

    return validate_configs(args.library, configs, args.show_payload)


if __name__ == "__main__":
    raise SystemExit(main())
