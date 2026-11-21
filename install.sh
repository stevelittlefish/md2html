#!/bin/sh
# Build md2html from source and install it onto your PATH (macOS and Linux).
#
#   ./install.sh                 installs to ~/.local/bin
#   ./install.sh /some/dir       installs to a directory of your choice
#
# Building locally means macOS Gatekeeper has nothing to quarantine: no
# notarisation, no "unidentified developer" prompts.
set -eu
cd "$(dirname "$0")"

DEST="${1:-$HOME/.local/bin}"

./build.sh

mkdir -p "$DEST"
install -m 0755 md2html "$DEST/md2html"

echo
echo "Installed $DEST/md2html"
case ":$PATH:" in
  *":$DEST:"*) "$DEST/md2html" --version ;;
  *)
    case "${SHELL:-}" in
      */zsh) rc="~/.zshrc" ;;
      */bash) rc="~/.bashrc" ;;
      *) rc="your shell's startup file" ;;
    esac
    echo "Note: $DEST is not on your PATH yet. Add this line to $rc, then open a new terminal:"
    echo "  export PATH=\"$DEST:\$PATH\""
    ;;
esac
