# Application releases

`.github/workflows/release.yml` builds version tags matching `vMAJOR.MINOR.PATCH`, or can be manually rerun on an existing tag. The tag must match `desktop/frontend/package.json`.

The `build` job uses native macOS ARM64, Windows x64, and Ubuntu 24.04 x64 runners. It calls `scripts/package-macos.sh`, `scripts/package-windows.ps1`, or `scripts/package-linux.sh deb`, reusing the existing packaging configuration and license inclusion. The macOS script ad-hoc signs the complete staged bundle after final plist updates, verifies its sealed resources, verifies the DMG, and mounts it read-only to verify the delivered bundle again. The DMG layout leaves the signed app’s Finder metadata intact (`scripts/dmg-settings.py` disables `hide_extensions`). Any failure stops publication. The workflow installs the DEB, and runs `go test -race ./...` on each target after frontend compilation.

`scripts/package-windows.ps1` retains the same built `net-test.exe` as `NET-Test-VERSION-windows-ARCH.exe` alongside the Setup installer, and copies `LICENSE` and `THIRD_PARTY_NOTICES.txt` to the output folder. The standalone download runs without NET-Test installation but still requires WebView2 Runtime. `desktop/main.go: main` and `desktop/internal/testengine/store.go: DefaultStorePath` / `OpenStore` continue to create and reuse `%APPDATA%\nt-wails\results.db`; no database ships with the executable.

Only after all builds and tests succeed does `publish` download the three installers, Windows standalone executable, and license files, verify the expected filenames, generate `SHA256SUMS.txt`, create a draft GitHub release, upload the packages, and publish it. Release notes come from `docs/architecture/release-notes.txt`; update them with future versions. Repeated runs replace assets for the same tag, so do not move published version tags.

No application APIs or runtime architecture change. RPM and Windows/Linux ARM64 are outside the current release matrix. Windows installers are unsigned. macOS apps are ad-hoc signed, without Developer ID signing or notarization; downloaded copies may require explicit approval. Interactive GUI smoke testing remains manual.

Starting with v2.0.1, the release matrix excludes macOS Intel. Local Intel packaging remains available through the existing macOS script.

The v2.0.1 macOS packaging correction replaces only the ARM64 DMG and its checksum in the existing release. The published v2.0.1 source tag remains unchanged; the packaging fix is committed separately on main and identified in the release notes.

The v2.0.1 Windows standalone addition extracts `net-test.exe` from the already published, checksum-verified Windows installer, preserving the exact released application build. It adds the standalone executable and accompanying license files and updates the checksum manifest and release notes. Existing installers and the v2.0.1 source tag remain unchanged.
