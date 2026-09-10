#!/usr/bin/env bash
set -euo pipefail

if [[ "${1:-}" == --help ]]; then
  echo 'Usage: bash scripts/package-linux.sh [deb|rpm] [version (default: app Info.plist)]'
  exit 0
fi
[[ $(uname -s) == Linux ]] || { echo 'Build Linux packages on Linux.' >&2; exit 1; }
ROOT=$(cd "$(dirname "$0")/.." && pwd)
FORMAT=${1:-deb}
VERSION=${2:-$(sed -n 's/.*<key>CFBundleShortVersionString<\/key><string>\([^<]*\)<\/string>.*/\1/p' "$ROOT/desktop/build/darwin/Info.plist")}
[[ $# -le 2 && "$FORMAT" =~ ^(deb|rpm)$ && "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || { echo 'Usage: package-linux.sh [deb|rpm] [major.minor.patch]' >&2; exit 1; }
for tool in go npm make pkg-config nfpm; do
  command -v "$tool" >/dev/null || { echo "Missing required tool: $tool (see BUILD_AND_INSTALL.md)." >&2; exit 1; }
done
[[ $(go env GOOS) == linux && $(go env GOARCH) == $(go env GOHOSTARCH) ]] || { echo 'Use the native Go OS and architecture (unset GOOS/GOARCH overrides).' >&2; exit 1; }
pkg-config --exists gtk4 webkitgtk-6.0 || { echo 'Install GTK4 and WebKitGTK 6.0 development libraries.' >&2; exit 1; }
export PACKAGE_ARCH=$(go env GOARCH)
export PACKAGE_VERSION="$VERSION"
mkdir -p "$ROOT/installation_package"
WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT
make -C "$ROOT/desktop" setup build
cd "$ROOT"
PACKAGE="$ROOT/installation_package/NET-Test-${VERSION}-linux-${PACKAGE_ARCH}.${FORMAT}"
nfpm package --config scripts/linux-package.yaml --packager "$FORMAT" --target "$WORK/package.$FORMAT"
mv -f "$WORK/package.$FORMAT" "$PACKAGE"
echo "Created: $PACKAGE"
