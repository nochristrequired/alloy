"""ctypes bindings for the Grafana Alloy shared parser library."""

from __future__ import annotations

import ctypes
import json
from dataclasses import dataclass
from enum import IntEnum
from pathlib import Path
from typing import Any, Dict, Optional


class ParseStatus(IntEnum):
    """Status codes returned by the shared library."""

    OK = 0
    WARNING = 1
    ERROR = 2
    INTERNAL_ERROR = 3


@dataclass
class ParseResult:
    """Container for a parsed payload and status information."""

    status: ParseStatus
    payload: Dict[str, Any]


class _AlloyJSONResult(ctypes.Structure):
    _fields_ = [
        ("data", ctypes.c_void_p),
        ("length", ctypes.c_size_t),
        ("status", ctypes.c_int),
    ]


class AlloyParserBridge:
    """Thin helper around the shared library entry points."""

    def __init__(self, library_path: Path | str) -> None:
        path = Path(library_path)
        if not path.exists():
            raise FileNotFoundError(path)

        self._lib = ctypes.CDLL(str(path))
        self._lib.AlloyParseFileJSON.argtypes = [
            ctypes.c_char_p,
            ctypes.c_void_p,
            ctypes.c_int,
        ]
        self._lib.AlloyParseFileJSON.restype = _AlloyJSONResult

        self._lib.AlloyParseExpressionJSON.argtypes = [
            ctypes.c_void_p,
            ctypes.c_int,
        ]
        self._lib.AlloyParseExpressionJSON.restype = _AlloyJSONResult

        self._lib.AlloyFree.argtypes = [ctypes.c_void_p]
        self._lib.AlloyFree.restype = None

    def parse_file(self, filename: str, data: bytes) -> ParseResult:
        """Parse an Alloy file and return its JSON payload."""

        filename_bytes = filename.encode("utf-8")
        data_ptr, length, keepalive = self._prepare_buffer(data)
        result = self._lib.AlloyParseFileJSON(filename_bytes, data_ptr, length)
        return self._consume_result(result, keepalive)

    def parse_expression(self, expression: bytes) -> ParseResult:
        """Parse a standalone Alloy expression."""

        data_ptr, length, keepalive = self._prepare_buffer(expression)
        result = self._lib.AlloyParseExpressionJSON(data_ptr, length)
        return self._consume_result(result, keepalive)

    def _prepare_buffer(
        self, payload: bytes
    ) -> tuple[ctypes.c_void_p, int, Optional[ctypes.Array[ctypes.c_char]]]:
        if not payload:
            return ctypes.c_void_p(), 0, None

        buffer = (ctypes.c_char * len(payload)).from_buffer_copy(payload)
        pointer = ctypes.cast(buffer, ctypes.c_void_p)
        return pointer, len(payload), buffer

    def _consume_result(
        self,
        result: _AlloyJSONResult,
        keepalive: Optional[ctypes.Array[ctypes.c_char]],
    ) -> ParseResult:
        _ = keepalive  # keep the buffer referenced until after parsing
        try:
            payload_bytes = b""
            if result.data and result.length:
                payload_bytes = ctypes.string_at(result.data, result.length)

            payload = json.loads(payload_bytes.decode("utf-8")) if payload_bytes else {}
        finally:
            if result.data:
                self._lib.AlloyFree(result.data)

        status = ParseStatus(result.status)
        return ParseResult(status=status, payload=payload)


__all__ = ["AlloyParserBridge", "ParseResult", "ParseStatus"]
