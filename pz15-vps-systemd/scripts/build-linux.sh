#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd -- "${SCRIPT_DIR}/.." && pwd)"

GOOS="${GOOS:-linux}"
GOARCH="${GOARCH:-amd64}"

cd "${PROJECT_DIR}"
mkdir -p bin

echo "Building tasks for ${GOOS}/${GOARCH}"
GOOS="${GOOS}" GOARCH="${GOARCH}" go build -o bin/tasks ./cmd/tasks
echo "Built ${PROJECT_DIR}/bin/tasks"
