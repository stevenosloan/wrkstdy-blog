#!/usr/bin/env bash
# Rebuilds the variable fonts in assets/fonts/. Only needed when changing which
# characters or weights ship; the output is committed.
#
#   Fira Sans: Firava (github.com/hellogreg/firava), a variable Fira Sans, cut
#              down to the latin range and to weights 300-500.
#   Fira Code: Fontsource's variable latin file, used as is (weights 300-700).
#
# Needs fontTools with brotli: pip install fonttools brotli
set -euo pipefail

FIRAVA_REF=8a8ed790c89ef661365c12d5f717f869d64c9c8f
FIRA_CODE_URL=https://cdn.jsdelivr.net/npm/@fontsource-variable/fira-code@5.3.0/files/fira-code-latin-wght-normal.woff2
WEIGHTS=300:500

# Google Fonts' "latin" subset, the same range Fontsource's latin files use.
LATIN=U+0000-00FF,U+0131,U+0152-0153,U+02BB-02BC,U+02C6,U+02DA,U+02DC,U+0304,U+0308,U+0329,U+2000-206F,U+20AC,U+2122,U+2191,U+2193,U+2212,U+2215,U+FEFF,U+FFFD

out="$(cd "$(dirname "$0")/.." && pwd)/assets/fonts"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

for style in normal italic; do
  src=Firava.woff2
  [ "$style" = italic ] && src=FiravaItalic.woff2
  curl -fsSL -o "$tmp/$src" "https://raw.githubusercontent.com/hellogreg/firava/$FIRAVA_REF/fonts/$src"
  pyftsubset "$tmp/$src" --unicodes="$LATIN" --layout-features='*' --flavor=woff2 --output-file="$tmp/latin-$src"
  fonttools varLib.instancer -q "$tmp/latin-$src" "wght=$WEIGHTS" -o "$out/firava-latin-wght-$style.woff2"
done

curl -fsSL -o "$out/fira-code-latin-wght-normal.woff2" "$FIRA_CODE_URL"

ls -l "$out"/*.woff2
