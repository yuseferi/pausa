#!/usr/bin/env bash

# Local release helper for Pausa.
#
# What it does:
#   1. Builds both macOS architectures with Wails
#   2. Packages each .app bundle into a zip
#   3. Prints SHA256 checksums for the Homebrew cask
#   4. Shows the exact next manual steps (update cask if needed, create tag,
#      push tag, publish or verify release)
#
# What it does NOT do automatically:
#   - commit version bumps
#   - update the cask file in-place
#   - create the GitHub release
#
# Those steps are intentionally left explicit so you can inspect the assets
# and checksums before publishing.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

if [[ $# -ne 1 ]]; then
  echo "usage: scripts/release.sh <version>" >&2
  echo "example: scripts/release.sh 1.0.2" >&2
  exit 1
fi

VERSION="$1"
TAG="v${VERSION}"

if ! command -v wails >/dev/null 2>&1; then
  echo "error: wails CLI not found in PATH" >&2
  exit 1
fi

echo "==> Preparing Pausa release ${TAG}"
echo

echo "==> Building frontend"
(cd frontend && npm install >/dev/null && npm run build)

rm -rf dist
mkdir -p dist

build_and_zip() {
  local platform="$1"
  local suffix="$2"

  echo
  echo "==> Building ${platform}"
  wails build -platform "${platform}"

  local zip="dist/pausa-${VERSION}-${suffix}.zip"
  echo "==> Packaging ${zip}"
  ditto -c -k --keepParent build/bin/pausa.app "$zip"
}

build_and_zip "darwin/arm64" "arm64-macos"
build_and_zip "darwin/amd64" "amd64-macos"

echo
echo "==> Checksums"
ARM_SHA=$(shasum -a 256 "dist/pausa-${VERSION}-arm64-macos.zip" | awk '{print $1}')
AMD_SHA=$(shasum -a 256 "dist/pausa-${VERSION}-amd64-macos.zip" | awk '{print $1}')

echo "arm64: ${ARM_SHA}"
echo "amd64: ${AMD_SHA}"

cat <<EOF

==> Next steps

1. Upload the two zip assets to a GitHub release named ${TAG}

   dist/pausa-${VERSION}-arm64-macos.zip
   dist/pausa-${VERSION}-amd64-macos.zip

2. Update Casks/pausa.rb to:

   version "${VERSION}"

   on_arm do
     sha256 "${ARM_SHA}"
     url "https://github.com/yuseferi/pausa/releases/download/v#{version}/pausa-#{version}-arm64-macos.zip"
   end

   on_intel do
     sha256 "${AMD_SHA}"
     url "https://github.com/yuseferi/pausa/releases/download/v#{version}/pausa-#{version}-amd64-macos.zip"
   end

3. Commit the version/cask changes

4. Tag and push:

   git tag ${TAG}
   git push origin main
   git push origin ${TAG}

If your GitHub Actions release workflow is enabled, pushing the tag will also
build and upload matching assets automatically.

EOF
