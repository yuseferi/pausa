#!/usr/bin/env bash
#
# Sign, notarize, staple, and package a macOS .app for Developer ID
# distribution. Used by .github/workflows/release.yml and available for local
# release builds.
#
# Usage:
#   scripts/sign-notarize.sh <path/to/App.app> <output.zip> [signing-identity]
#
# Signing:
#   IDENTITY / 3rd arg    "Developer ID Application: Name (TEAMID)". If omitted,
#                         the first Developer ID Application identity in the
#                         keychain is used.
#   ENTITLEMENTS          Default: build/darwin/entitlements.plist
#
# Notarization credentials (provide ONE of):
#   API key:   NOTARY_KEY_PATH (path to .p8), NOTARY_KEY_ID, [NOTARY_ISSUER]
#   Apple ID:  APPLE_ID, APPLE_APP_PASSWORD, APPLE_TEAM_ID
#
# Other:
#   NOTARY_TIMEOUT        Default: 30m
#
# The archive is produced with `ditto` AFTER stapling, because the notarization
# ticket is stored in the bundle's extended attributes; a plain `zip` would
# strip it.
set -euo pipefail

APP="${1:?usage: sign-notarize.sh <App.app> <output.zip> [identity]}"
OUT_ZIP="${2:?usage: sign-notarize.sh <App.app> <output.zip> [identity]}"
IDENTITY="${3:-${IDENTITY:-}}"
ENTITLEMENTS="${ENTITLEMENTS:-build/darwin/entitlements.plist}"
NOTARY_TIMEOUT="${NOTARY_TIMEOUT:-30m}"

[[ -d "$APP" ]] || { echo "error: app bundle not found: $APP" >&2; exit 1; }
[[ -f "$ENTITLEMENTS" ]] || { echo "error: entitlements not found: $ENTITLEMENTS" >&2; exit 1; }

# --- Resolve signing identity -------------------------------------------------
if [[ -z "$IDENTITY" ]]; then
  IDENTITY="$(security find-identity -v -p codesigning \
    | sed -n 's/.*"\(Developer ID Application:.*\)"/\1/p' | head -1)"
fi
[[ -n "$IDENTITY" ]] || { echo "error: no 'Developer ID Application' identity found" >&2; exit 1; }
echo "==> Signing with: $IDENTITY"

# --- Sign ---------------------------------------------------------------------
# A Wails app contains a single Mach-O executable and no nested frameworks or
# helpers, so signing the bundle (which seals the executable) is sufficient.
# Never use --deep to sign; it is deprecated and signs the wrong things.
codesign --force --options runtime --timestamp \
  --entitlements "$ENTITLEMENTS" \
  --sign "$IDENTITY" \
  "$APP"

echo "==> Verifying signature"
codesign --verify --deep --strict --verbose=2 "$APP"
codesign -dvv "$APP" 2>&1 | grep -qi '^Timestamp=' \
  || { echo "error: signature is missing a secure timestamp" >&2; exit 1; }

# --- Notarize -----------------------------------------------------------------
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

SUBMIT_ZIP="$WORK/notarize.zip"
echo "==> Creating notarization archive"
ditto -c -k --keepParent "$APP" "$SUBMIT_ZIP"

NOTARY_ARGS=()
if [[ -n "${NOTARY_KEY_PATH:-}" ]]; then
  NOTARY_ARGS+=(--key "$NOTARY_KEY_PATH" --key-id "${NOTARY_KEY_ID:?NOTARY_KEY_ID required with NOTARY_KEY_PATH}")
  # --issuer is required for Team API keys and must be omitted for Individual keys.
  [[ -n "${NOTARY_ISSUER:-}" ]] && NOTARY_ARGS+=(--issuer "$NOTARY_ISSUER")
elif [[ -n "${APPLE_ID:-}" ]]; then
  NOTARY_ARGS+=(--apple-id "$APPLE_ID" \
    --password "${APPLE_APP_PASSWORD:?APPLE_APP_PASSWORD required}" \
    --team-id "${APPLE_TEAM_ID:?APPLE_TEAM_ID required}")
else
  echo "error: set NOTARY_KEY_PATH/NOTARY_KEY_ID (API key) or APPLE_ID/APPLE_APP_PASSWORD/APPLE_TEAM_ID (Apple ID)" >&2
  exit 1
fi

echo "==> Submitting for notarization (timeout ${NOTARY_TIMEOUT})"
set +e
SUBMIT_JSON="$(xcrun notarytool submit "$SUBMIT_ZIP" "${NOTARY_ARGS[@]}" \
  --wait --timeout "$NOTARY_TIMEOUT" --output-format json)"
RC=$?
set -e

SUBMISSION_ID="$(printf '%s' "$SUBMIT_JSON" | python3 -c 'import sys,json; print(json.load(sys.stdin).get("id",""))' 2>/dev/null || true)"
if [[ $RC -ne 0 || -z "$SUBMISSION_ID" ]]; then
  echo "error: notarization failed (exit $RC)" >&2
  printf '%s\n' "$SUBMIT_JSON" >&2
  if [[ -n "$SUBMISSION_ID" ]]; then
    echo "==> Notarization log" >&2
    xcrun notarytool log "$SUBMISSION_ID" "${NOTARY_ARGS[@]}" >&2 || true
  fi
  exit 1
fi
echo "==> Notarization accepted ($SUBMISSION_ID)"

# --- Staple + Gatekeeper assessment -------------------------------------------
echo "==> Stapling ticket"
xcrun stapler staple "$APP"
xcrun stapler validate "$APP"

echo "==> Gatekeeper assessment"
spctl --assess --type exec --verbose=4 "$APP"

# --- Package ------------------------------------------------------------------
echo "==> Packaging $OUT_ZIP"
mkdir -p "$(dirname "$OUT_ZIP")"
rm -f "$OUT_ZIP"
ditto -c -k --keepParent "$APP" "$OUT_ZIP"
shasum -a 256 "$OUT_ZIP" | awk '{print $1}' > "$OUT_ZIP.sha256"
echo "==> sha256: $(cat "$OUT_ZIP.sha256")"
echo "==> Done: signed, notarized, stapled"
