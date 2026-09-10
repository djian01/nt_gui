#!/usr/bin/env bash
set -euo pipefail

if [[ "${1:-}" == --help ]]; then
  echo 'Usage: bash scripts/package-macos.sh [version (default: app Info.plist)]'
  exit 0
fi
[[ $(uname -s) == Darwin ]] || { echo 'Build macOS packages on macOS.' >&2; exit 1; }
ROOT=$(cd "$(dirname "$0")/.." && pwd)
# Prefer the documented local environment without changing the caller's PATH.
DMGBUILD="$ROOT/.local/packaging-venv/bin/dmgbuild"
if [[ ! -x "$DMGBUILD" ]]; then DMGBUILD=dmgbuild; fi
for tool in go npm make hdiutil "$DMGBUILD" swift; do
  command -v "$tool" >/dev/null || { echo "Missing required tool: $tool (see BUILD_AND_INSTALL.md)." >&2; exit 1; }
done
VERSION=${1:-$(/usr/libexec/PlistBuddy -c 'Print :CFBundleShortVersionString' "$ROOT/desktop/build/darwin/Info.plist")}
[[ $# -le 1 && "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || { echo 'Version must be major.minor.patch.' >&2; exit 1; }
[[ $(go env GOOS) == darwin && $(go env GOARCH) == $(go env GOHOSTARCH) ]] || { echo 'Use the native Go OS and architecture (unset GOOS/GOARCH overrides).' >&2; exit 1; }
ARCH=$(go env GOARCH)
mkdir -p "$ROOT/installation_package"
WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT
make -C "$ROOT/desktop" setup mac-app
mkdir -p "$WORK/image"
ditto "$ROOT/desktop/bin/NET-Test.app" "$WORK/image/NET-Test.app"
/usr/libexec/PlistBuddy -c "Set :CFBundleShortVersionString $VERSION" "$WORK/image/NET-Test.app/Contents/Info.plist"
/usr/libexec/PlistBuddy -c "Set :CFBundleVersion $VERSION" "$WORK/image/NET-Test.app/Contents/Info.plist"
PACKAGE="$ROOT/installation_package/NET-Test-${VERSION}-macos-${ARCH}.dmg"
"$DMGBUILD" -s "$ROOT/scripts/dmg-settings.py" \
  -D "app=$WORK/image/NET-Test.app" \
  -D "assets=$ROOT/scripts/installer-assets" 'NET-Test Installer' "$WORK/package.dmg"
swift "$ROOT/scripts/set-macos-file-icon.swift" "$ROOT/scripts/installer-assets/installer.icns" "$WORK/package.dmg"
mv -f "$WORK/package.dmg" "$PACKAGE"
echo "Created: $PACKAGE"
echo 'This package is unsigned and not notarized; public distribution requires Apple signing and notarization.'
