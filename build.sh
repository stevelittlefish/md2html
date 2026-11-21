#!/bin/sh
# One-command build for macOS and Linux.
# Uses the system Go if it is new enough, otherwise downloads the official
# Go toolchain into ./.toolchain (no sudo, nothing installed system-wide).
set -eu

GO_VERSION="1.26.4"
MIN_MINOR=24
cd "$(dirname "$0")"

have_go() {
  command -v go >/dev/null 2>&1 || return 1
  v=$(go env GOVERSION 2>/dev/null | sed 's/^go//')
  minor=$(echo "$v" | cut -d. -f2)
  [ "$(echo "$v" | cut -d. -f1)" -ge 1 ] && [ "${minor:-0}" -ge "$MIN_MINOR" ]
}

if have_go; then
  GO=go
else
  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  case "$(uname -m)" in
    x86_64|amd64) arch=amd64 ;;
    arm64|aarch64) arch=arm64 ;;
    *) echo "Unsupported CPU: $(uname -m)" >&2; exit 1 ;;
  esac
  GO="$PWD/.toolchain/go/bin/go"
  if [ ! -x "$GO" ]; then
    url="https://go.dev/dl/go${GO_VERSION}.${os}-${arch}.tar.gz"
    echo "Go not found; downloading $url"
    mkdir -p .toolchain
    curl -fsSL "$url" | tar -xz -C .toolchain
  fi
fi

version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)
CGO_ENABLED=0 "$GO" build -trimpath -ldflags "-s -w -X main.version=$version" -o md2html .
echo "Built $PWD/md2html ($version)"
echo "Install it with:  sudo mv md2html /usr/local/bin/   (or anywhere on your PATH)"
