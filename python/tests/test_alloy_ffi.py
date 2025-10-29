"""Integration tests for the Alloy parser shared library bridge."""

from __future__ import annotations

import ctypes
import json
import subprocess
from pathlib import Path
import sys

import pytest

PROJECT_ROOT = Path(__file__).resolve().parents[2]
if str(PROJECT_ROOT) not in sys.path:
    sys.path.insert(0, str(PROJECT_ROOT))

from python import alloy_ffi as alloy_ffi_module
from python.alloy_ffi import AlloyParserBridge, ParseStatus


@pytest.fixture(scope="session")
def shared_library(tmp_path_factory: pytest.TempPathFactory) -> Path:
    """Build the shared library once for all tests."""

    build_dir = tmp_path_factory.mktemp("alloy_ffi")
    library_path = build_dir / "liballoyparser.so"
    syntax_root = PROJECT_ROOT / "syntax"

    subprocess.run(
        [
            "go",
            "build",
            "-buildmode=c-shared",
            "-o",
            str(library_path),
            "./ffi",
        ],
        check=True,
        cwd=syntax_root,
    )

    return library_path


@pytest.fixture()
def bridge(shared_library: Path) -> AlloyParserBridge:
    return AlloyParserBridge(shared_library)


def test_bridge_requires_existing_library(tmp_path: Path) -> None:
    missing = tmp_path / "missing.so"
    with pytest.raises(FileNotFoundError):
        AlloyParserBridge(missing)


def test_parse_file_success(bridge: AlloyParserBridge) -> None:
    source = b"""// leading comment\nblock "label" {\n  nested = [1, 2, 3]\n}\nattribute = "value"\n"""

    result = bridge.parse_file("example.alloy", source)

    assert result.status == ParseStatus.OK
    payload = result.payload
    assert payload["schemaVersion"] == 1

    file_section = payload["file"]
    assert file_section["name"] == "example.alloy"
    assert len(file_section["body"]) == 2

    block_stmt = file_section["body"][0]
    assert block_stmt["kind"] == "Block"
    block = block_stmt["block"]
    assert block["name"] == ["block"]
    assert block["label"] == "label"
    assert len(block["body"]) == 1

    nested_stmt = block["body"][0]
    assert nested_stmt["kind"] == "Attribute"
    array_expr = nested_stmt["attribute"]["value"]
    assert array_expr["kind"] == "Array"
    assert [item["literal"]["value"] for item in array_expr["array"]["elements"]] == ["1", "2", "3"]

    attribute_stmt = file_section["body"][1]
    assert attribute_stmt["kind"] == "Attribute"
    assert attribute_stmt["attribute"]["name"]["name"] == "attribute"
    literal = attribute_stmt["attribute"]["value"]["literal"]
    assert literal["token"] == "STRING"
    assert literal["value"] == '"value"'

    comments = file_section["comments"]
    assert comments and comments[0]["comments"][0]["text"].strip() == "// leading comment"

    assert payload["diagnostics"] == []


def test_parse_file_with_error(bridge: AlloyParserBridge) -> None:
    result = bridge.parse_file("broken.alloy", b"attr =\n")

    assert result.status == ParseStatus.ERROR
    diagnostics = result.payload["diagnostics"]
    assert diagnostics, "Expected diagnostics for invalid configuration"
    severities = {entry["severity"] for entry in diagnostics}
    assert "error" in severities


def test_parse_expression_success(bridge: AlloyParserBridge) -> None:
    result = bridge.parse_expression(b"(1 + 2) * 3")

    assert result.status == ParseStatus.OK
    expr = result.payload["expression"]
    assert expr["kind"] == "Binary"
    binary = expr["binary"]
    assert binary["operator"] == "*"
    left = binary["left"]
    assert left["kind"] == "Paren"
    assert left["paren"]["inner"]["binary"]["operator"] == "+"
    assert result.payload["diagnostics"] == []


def test_parse_expression_with_error(bridge: AlloyParserBridge) -> None:
    result = bridge.parse_expression(b"1 +")

    assert result.status == ParseStatus.ERROR
    diagnostics = result.payload["diagnostics"]
    assert diagnostics
    assert any(entry["severity"] == "error" for entry in diagnostics)


def test_results_are_valid_json(bridge: AlloyParserBridge) -> None:
    result = bridge.parse_file("simple.alloy", b"foo = 1\n")
    raw_payload = json.dumps(result.payload)
    decoded = json.loads(raw_payload)
    assert decoded["file"]["body"][0]["attribute"]["value"]["literal"]["value"] == "1"


def test_parse_expression_large_length(
    monkeypatch: pytest.MonkeyPatch, bridge: AlloyParserBridge
) -> None:
    """Guard against 32-bit truncation of large configurations."""

    large_length = 2**31 + 1
    sentinel_pointer = ctypes.c_void_p(0xDEADBEEF)

    def fake_prepare_buffer(_: bytes) -> tuple[ctypes.c_void_p, ctypes.c_size_t, None]:
        return sentinel_pointer, ctypes.c_size_t(large_length), None

    call_count = 0

    def fake_parse_expression(
        data_ptr: ctypes.c_void_p, length: ctypes.c_size_t
    ) -> alloy_ffi_module._AlloyJSONResult:
        nonlocal call_count
        call_count += 1
        assert isinstance(length, ctypes.c_size_t)
        assert length.value == large_length
        assert data_ptr.value == sentinel_pointer.value
        return alloy_ffi_module._AlloyJSONResult(
            data=None,
            length=0,
            status=ParseStatus.OK.value,
        )

    monkeypatch.setattr(bridge, "_prepare_buffer", fake_prepare_buffer)
    monkeypatch.setattr(bridge._lib, "AlloyParseExpressionJSON", fake_parse_expression)

    result = bridge.parse_expression(b"not actually parsed")

    assert result.status == ParseStatus.OK
    assert result.payload == {}
    assert call_count == 1
