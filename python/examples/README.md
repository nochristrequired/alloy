# Validate Alloy configuration examples

This example shows you how to build the Go-based Alloy parser as a shared library and validate configuration files from Python.

Before you begin, ensure you have the following:

- Bash 5 or later
- Go 1.25 or later
- Python 3.11 or later

## Build the parser library

Run the helper script to compile `liballoyparser` into the `dist/` directory:

```sh
./python/examples/build_liballoyparser.sh
```

The script verifies that `go` and `make` are available before invoking the Makefile target. It prints the absolute path to the compiled shared library.

## Validate example configurations

Run the validation wrapper to check the bundled Alloy configuration files:

```sh
./python/examples/run_validation_example.sh
```

The script locates the shared library under `dist/` by default. Pass a different library path as the first argument to validate the files against an alternative build:

```sh
./python/examples/run_validation_example.sh /custom/path/liballoyparser.so
```

Provide additional configuration files after the optional library argument to validate more Alloy configurations in one run.

## Run inside Docker

Use the Docker helpers if you prefer to run the validation workflow without installing Go or Python locally.

1. Build the container image (optionally pass a custom tag as the first argument):

   ```sh
   ./python/examples/docker_build_image.sh
   ```

2. Execute the build-and-validate sequence inside the container. The script accepts an optional image tag followed by the same arguments that `run_validation_example.sh` supports:

   ```sh
   ./python/examples/docker_run_example.sh
   ```

   For example, to validate an additional configuration file:

   ```sh
   ./python/examples/docker_run_example.sh alloy-parser-example:latest python/examples/alloy_config_valid.alloy
   ```

Both scripts mount the repository into the container so that `docker_run_in_container.sh` can reuse the existing build and validation helpers.

## Next steps

- Review the [`run_validation.py`](./run_validation.py) helper for more invocation options.
- Explore the [`python/alloy_ffi.py`](../alloy_ffi.py) bridge to integrate the parser into other Python tooling.
