# Installer releases

`.github/workflows/release.yml` builds version tags matching `vMAJOR.MINOR.PATCH`, or can be manually rerun on an existing tag. The tag must match `desktop/frontend/package.json`.

The `build` job uses native macOS ARM64, Windows x64, and Ubuntu 24.04 x64 runners. It calls `scripts/package-macos.sh`, `scripts/package-windows.ps1`, or `scripts/package-linux.sh deb`, reusing the existing packaging configuration and license inclusion. It verifies DMGs, installs the DEB, and runs `go test -race ./...` on each target after frontend compilation.

Only after all builds and tests succeed does `publish` download the three installers, generate `SHA256SUMS.txt`, create a draft GitHub release, upload the packages, and publish it. Release notes come from `docs/architecture/release-notes.txt`; update them with future versions. Repeated runs replace assets for the same tag, so do not move published version tags.

No application APIs or runtime architecture change. RPM and Windows/Linux ARM64 are outside the current release matrix. Installers are unsigned and macOS packages are not notarized. Interactive GUI smoke testing remains manual.

Starting with v2.0.1, the release matrix excludes macOS Intel. Local Intel packaging remains available through the existing macOS script.
