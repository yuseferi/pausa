#!/usr/bin/env bash
#
# One-line installer for Pausa (unsigned macOS build).
#
#   curl -fsSL https://raw.githubusercontent.com/yuseferi/pausa/main/scripts/install.sh | bash
#
# Environment options:
#   PAUSA_VERSION   install a specific version (e.g. 1.4.1). Default: latest release.
#   PAUSA_DEST      install directory. Default: /Applications (falls back to ~/Applications).
#   PAUSA_NO_OPEN   set to 1 to skip launching the app after install.
#
# Pausa is distributed unsigned (there is no paid Apple Developer account), so a
# downloaded bundle carries the com.apple.quarantine attribute and macOS refuses
# the first launch. This script strips that attribute after installing so the
# app opens without the manual `xattr -dr` step.

set -euo pipefail

REPO="yuseferi/pausa"
APP_NAME="pausa"

log() { printf '==> %s\n' "$*"; }
die() { printf 'error: %s\n' "$*" >&2; exit 1; }

[[ "$(uname -s)" == "Darwin" ]] || die "Pausa is macOS-only (detected $(uname -s))."

case "$(uname -m)" in
  arm64)  SUFFIX="arm64-macos" ;;
  x86_64) SUFFIX="amd64-macos" ;;
  *)      die "unsupported architecture: $(uname -m)" ;;
esac

for tool in curl shasum ditto unzip xattr; do
  command -v "$tool" >/dev/null 2>&1 || die "missing required tool: $tool"
done

# --- Resolve version --------------------------------------------------------
VERSION="${PAUSA_VERSION:-}"
if [[ -z "$VERSION" ]]; then
  log "Finding the latest Pausa release…"
  VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep -m1 '"tag_name"' \
    | sed -E 's/.*"tag_name": *"v?([^"]+)".*/\1/')" || true
fi
[[ -n "$VERSION" ]] || die "could not determine the latest version; set PAUSA_VERSION=x.y.z and retry"

ASSET="pausa-${VERSION}-${SUFFIX}.zip"
BASE_URL="https://github.com/${REPO}/releases/download/v${VERSION}"

# --- Resolve destination ----------------------------------------------------
DEST="${PAUSA_DEST:-/Applications}"
if ! mkdir -p "$DEST" 2>/dev/null || [[ ! -w "$DEST" ]]; then
  DEST="${HOME}/Applications"
  log "${PAUSA_DEST:-/Applications} is not writable; installing to ${DEST}"
  mkdir -p "$DEST"
fi

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

# --- Download ---------------------------------------------------------------
log "Downloading ${ASSET}"
curl -fL --proto '=https' --tlsv1.2 -o "$TMP/$ASSET" "${BASE_URL}/${ASSET}" \
  || die "download failed: ${BASE_URL}/${ASSET}"

# --- Verify checksum (when the release publishes one) -----------------------
if curl -fsSL -o "$TMP/$ASSET.sha256" "${BASE_URL}/${ASSET}.sha256"; then
  expected="$(tr -d '[:space:]' < "$TMP/$ASSET.sha256")"
  actual="$(shasum -a 256 "$TMP/$ASSET" | awk '{print $1}')"
  [[ "$expected" == "$actual" ]] || die "checksum mismatch (expected ${expected}, got ${actual})"
  log "Checksum verified"
else
  log "warning: no checksum published for ${ASSET}; skipping verification"
fi

# --- Install ----------------------------------------------------------------
log "Installing to ${DEST}/${APP_NAME}.app"
rm -rf "$DEST/${APP_NAME}.app"
ditto -x -k "$TMP/$ASSET" "$TMP/extract"

if [[ -d "$TMP/extract/${APP_NAME}.app" ]]; then
  mv "$TMP/extract/${APP_NAME}.app" "$DEST/${APP_NAME}.app"
else
  # The release zip normally keeps the bundle at the top level (ditto
  # --keepParent), but fall back to the first .app found just in case.
  found="$(find "$TMP/extract" -maxdepth 2 -name '*.app' -print -quit)"
  [[ -n "$found" ]] || die "no .app bundle found inside ${ASSET}"
  mv "$found" "$DEST/${APP_NAME}.app"
fi

# Unsigned build: clear the quarantine bit so the first launch is not blocked.
xattr -dr com.apple.quarantine "$DEST/${APP_NAME}.app" 2>/dev/null || true

log "Installed Pausa ${VERSION}"

if [[ "${PAUSA_NO_OPEN:-0}" != "1" ]]; then
  log "Launching Pausa…"
  open "$DEST/${APP_NAME}.app" || true
fi
