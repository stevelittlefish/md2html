#!/bin/sh
# One-command build for macOS and Linux: produces ./md2html.
# Uses the system Go if it is new enough for go.mod, otherwise downloads the
# official Go toolchain into ./.toolchain (no sudo, nothing installed system-wide).
set -eu
cd "$(dirname "$0")"

# The Go version go.mod asks for, e.g. 1.26.4.
GO_VERSION=$(sed -n 's/^go \([0-9.]*\)$/\1/p' go.mod)

# Succeeds if version $1 >= version $2 (dotted numbers).
version_ge() {
  [ "$(printf '%s\n%s\n' "$2" "$1" | sort -t. -k1,1n -k2,2n -k3,3n | head -n1)" = "$2" ]
}

GO=""
if command -v go >/dev/null 2>&1; then
  have=$(GOTOOLCHAIN=local go env GOVERSION 2>/dev/null | sed 's/^go//; s/[^0-9.].*//')
  if version_ge "$have" "$GO_VERSION"; then GO=go; fi
fi

if [ -z "$GO" ]; then
  GO="$PWD/.toolchain/go/bin/go"
  if [ ! -x "$GO" ] || ! version_ge "$(GOTOOLCHAIN=local "$GO" env GOVERSION | sed 's/^go//')" "$GO_VERSION"; then
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    case "$(uname -m)" in
      x86_64|amd64) arch=amd64 ;;
      arm64|aarch64) arch=arm64 ;;
      *) echo "Unsupported CPU: $(uname -m)" >&2; exit 1 ;;
    esac
    url="https://go.dev/dl/go${GO_VERSION}.${os}-${arch}.tar.gz"
    echo "Downloading Go $GO_VERSION from $url"
    rm -rf .toolchain && mkdir -p .toolchain
    curl -fsSL "$url" | tar -xz -C .toolchain
  fi
fi

# Release archives carry a VERSION file; git checkouts use git describe.
if [ -f VERSION ]; then
  version=$(cat VERSION)
else
  version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)
fi

GOTOOLCHAIN=local CGO_ENABLED=0 "$GO" build -trimpath -ldflags "-s -w -X main.version=$version" -o md2html .
echo "Built $PWD/md2html ($version)"
