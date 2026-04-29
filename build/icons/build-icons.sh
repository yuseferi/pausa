#!/usr/bin/env bash
# Generate the macOS .icns file from build/icons/appicon.svg, plus the
# template PNGs for the menu-bar status item from statusbar.svg.
#
# Outputs:
#   build/appicon.png            -- 1024x1024 PNG (Wails app icon source)
#   build/icons/appicon.iconset/ -- all sizes Apple expects
#   build/icons/appicon.icns     -- assembled .icns (used by .app bundle)
#   build/icons/menubar/         -- statusbar template PNGs (1x/2x/3x)
#
# Requires only macOS built-ins: qlmanage, sips, iconutil.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
APP_SVG="${SCRIPT_DIR}/appicon.svg"
BAR_SVG="${SCRIPT_DIR}/statusbar.svg"
ICONSET="${SCRIPT_DIR}/appicon.iconset"
ICNS_OUT="${SCRIPT_DIR}/appicon.icns"
MENU_DIR="${SCRIPT_DIR}/menubar"

if [[ ! -f "${APP_SVG}" ]]; then
  echo "missing ${APP_SVG}" >&2
  exit 1
fi

# render_svg <svg> <size> <output_png>
# Renders the SVG at the requested size using qlmanage (QuickLook), then
# uses sips to crop to exact dimensions.
render_svg() {
  local svg="$1" size="$2" out="$3"
  local tmpdir
  tmpdir="$(mktemp -d)"
  qlmanage -t -s "${size}" -o "${tmpdir}" "${svg}" >/dev/null 2>&1
  local rendered="${tmpdir}/$(basename "${svg}").png"
  if [[ ! -f "${rendered}" ]]; then
    echo "qlmanage failed for ${svg} @${size}" >&2
    rm -rf "${tmpdir}"
    exit 1
  fi
  # qlmanage may pad to a square; force exact size with sips.
  sips -z "${size}" "${size}" "${rendered}" --out "${out}" >/dev/null
  rm -rf "${tmpdir}"
}

echo "==> Building app icon (.iconset + .icns)"
rm -rf "${ICONSET}"
mkdir -p "${ICONSET}"

# Apple's required sizes for a Universal .icns
#   16, 32, 64, 128, 256, 512, 1024 (each in 1x and 2x except 1024)
declare -a SPECS=(
  "16:icon_16x16.png"
  "32:icon_16x16@2x.png"
  "32:icon_32x32.png"
  "64:icon_32x32@2x.png"
  "128:icon_128x128.png"
  "256:icon_128x128@2x.png"
  "256:icon_256x256.png"
  "512:icon_256x256@2x.png"
  "512:icon_512x512.png"
  "1024:icon_512x512@2x.png"
)
for spec in "${SPECS[@]}"; do
  size="${spec%%:*}"
  name="${spec##*:}"
  echo "    ${name} (${size}x${size})"
  render_svg "${APP_SVG}" "${size}" "${ICONSET}/${name}"
done

echo "==> Assembling ${ICNS_OUT}"
iconutil -c icns "${ICONSET}" -o "${ICNS_OUT}"

# Wails consumes a 1024x1024 PNG named build/appicon.png; mirror our master.
echo "==> Updating ${BUILD_DIR}/appicon.png (1024x1024)"
render_svg "${APP_SVG}" 1024 "${BUILD_DIR}/appicon.png"

echo "==> Building menu-bar template PNGs"
mkdir -p "${MENU_DIR}"
# Standard menu-bar logical size is 22pt. Provide 1x/2x/3x retina assets.
declare -a MENU_SIZES=(
  "22:menubar.png"
  "44:menubar@2x.png"
  "66:menubar@3x.png"
)
for spec in "${MENU_SIZES[@]}"; do
  size="${spec%%:*}"
  name="${spec##*:}"
  echo "    ${name} (${size}x${size})"
  render_svg "${BAR_SVG}" "${size}" "${MENU_DIR}/${name}"
done

echo
echo "Done."
echo "  - App icon:      ${ICNS_OUT}"
echo "  - Wails source:  ${BUILD_DIR}/appicon.png"
echo "  - Menubar PNGs:  ${MENU_DIR}/"
