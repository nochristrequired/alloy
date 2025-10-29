# syntax=docker/dockerfile:1
FROM python:3.11-slim AS base

ARG GO_VERSION=1.25.1

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        git \
        make \
        build-essential \
    && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o /tmp/go.tgz \
    && tar -C /usr/local -xzf /tmp/go.tgz \
    && rm /tmp/go.tgz

ENV PATH="/usr/local/go/bin:${PATH}"

WORKDIR /workspace

ENTRYPOINT ["/bin/bash"]
