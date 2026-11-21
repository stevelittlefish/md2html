#!/bin/sh
# Build md2html from source and install it onto your PATH (macOS and Linux).
#
#   ./install.sh                 installs to /usr/local/bin (asks for sudo if needed)
#   ./install.sh ~/.local/bin    installs to a directory of your choice
#
# Building locally means macOS Gatekeeper has nothing to quarantine: no
# notarisation, no "unidentified developer" prompts.
set -eu
cd "$(dirname "$0")"

DEST="${1:-/usr/local/bin}"

./build.sh

mkdir -p "$DEST" 2>/dev/null || sudo mkdir -p "$DEST"
if [ -w "$DEST" ]; then
  install -m 0755 md2html "$DEST/md2html"
else
  echo "Installing to $DEST needs administrator rights."
  sudo install -m 0755 md2html "$DEST/md2html"
fi

echo
echo "Installed $DEST/md2html"
case ":$PATH:" in
  *":$DEST:"*) "$DEST/md2html" --version ;;
  *) echo "Note: $DEST is not on your PATH. Add it, e.g.:  export PATH=\"$DEST:\$PATH\"" ;;
esac
