#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

if ! command -v go >/dev/null 2>&1; then
  echo "go not found on PATH"
  exit 1
fi

rm -rf dist
mkdir -p dist

build_target() {
  local goos="$1"
  local goarch="$2"
  local out_dir="$3"
  local out_file="$4"
  local ldflags="-s -w"

  mkdir -p "dist/$out_dir"
  echo "Building $goos/$goarch -> dist/$out_dir/$out_file"

  if [[ -n "${GC_CLI_DEFAULT_OAUTH_CLIENT_SECRET:-}" ]]; then
    ldflags="$ldflags -X github.com/timothy/gc-cli/internal/auth.DefaultOAuthClientSecret=${GC_CLI_DEFAULT_OAUTH_CLIENT_SECRET}"
  fi

  CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" \
    go build -trimpath -ldflags="$ldflags" -o "dist/$out_dir/$out_file" ./cmd/gc
}

build_target darwin arm64 darwin-arm64 gc-cli
build_target darwin amd64 darwin-amd64 gc-cli
build_target linux arm64 linux-arm64 gc-cli
build_target linux amd64 linux-amd64 gc-cli
build_target windows arm64 windows-arm64 gc-cli.exe
build_target windows amd64 windows-amd64 gc-cli.exe

chmod +x dist/darwin-arm64/gc-cli \
  dist/darwin-amd64/gc-cli \
  dist/linux-arm64/gc-cli \
  dist/linux-amd64/gc-cli

echo "Done."
