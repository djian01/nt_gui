# Net Test branding and platform icons

## Source and rendering

`desktop/frontend/public/net-test.svg` is the editable master: a white N-shaped
network route with two endpoint rings on the existing royal-blue accent
(`#2864D8`). Flat geometry and transparent outer corners keep it readable in
both themes and at small icon sizes. Endpoint fills cover the route beneath
them so their counters remain open.

`desktop/frontend/src/App.tsx: App` displays this SVG beside a title-case
Net Test wordmark. `desktop/frontend/src/style.css: .brand-mark` preserves its
square proportions; `.brand strong` styles the wordmark. The image is
decorative because the adjacent text supplies the accessible app name.
`desktop/frontend/index.html` uses the same SVG as the favicon.

`desktop/frontend/scripts/generate-icons.mjs` renders the vector using pinned
Sharp through `npm --prefix frontend run icons` / `make icons`. Its top-level
generation code writes:

| Asset | Purpose |
| --- | --- |
| `Icon.png` | 1024-pixel README and legacy Fyne icon |
| `desktop/build/icons/net-test.png` | 512-pixel Wails runtime and Linux fallback icon |
| `desktop/build/icons/net-test.ico` | Windows 16, 24, 32, 48, 64, 128, and 256-pixel PNG-backed icon entries |
| `desktop/build/icons/net-test.icns` | macOS standard and Retina entries up to 1024 pixels |

Generated icons are application assets and belong in Git. Regeneration is an
explicit design step, not work repeated on every build. Temporary Windows
resource objects are ignored. Local project notes remain in ignored
`docs/memory/`; permanent documentation belongs here.

## Native integration

`desktop/main.go: main` passes the embedded `appIcon` PNG to
`application.New(application.Options{Icon: ...})` for runtime windows/About
branding. The pinned Wails Windows window code falls back to this PNG when
its preferred numbered icon resource is absent.

- macOS: `desktop/Makefile: mac-app` copies `net-test.icns` into the bundle's
  `Contents/Resources`. `desktop/build/darwin/Info.plist: CFBundleIconFile`
  points to it. Rebuilding the plain executable does not refresh the bundle.
- Windows: `desktop/Makefile: windows-icons` runs pinned
  `github.com/akavel/rsrc@v0.10.2` with the target `GOARCH`, creating
  `rsrc_windows_<arch>.syso`; Windows `build` depends on this target. The Go
  linker includes the object in the executable. PowerShell and CI run the same
  generation command before compiling. The PNG-backed ICO format and resource
  embedding follow the [rsrc tool](https://github.com/akavel/rsrc).
- Linux: `desktop/build/linux/net-test.desktop` references the `net-test`
  themed icon and executable, following the
  [Desktop Entry Specification](https://specifications.freedesktop.org/desktop-entry/latest/).
  `desktop/Makefile: linux-package` stages an installable `usr/` tree with the
  binary, launcher, scalable SVG, and 512-pixel PNG. The build guide provides
  user-local installation with an absolute launcher executable path. ELF files
  themselves do not supply desktop file-manager icons.

`.github/workflows/desktop.yml` packages a macOS app ZIP, Windows EXE, and Linux
tar archive containing desktop integration files. The archive formats preserve
Unix permissions. CI validates the Linux desktop entry.

Legacy compatibility: `Resource_Shared_png.go` embeds `Icon.png` directly and
keeps `resourceIconPng` for the existing About window; `main.go: main` also calls
`a.SetIcon(resourceIconPng)`. This removes the obsolete bundled byte literal.

## Scope and limitations

No service API, database, test behavior, or authentication setting changed.
Native packaging still requires the target operating system's Wails toolchain;
cross-generated icon resources alone are not evidence of native runtime
compatibility. macOS bundles remain unsigned. Installed copies must be replaced
after rebuilding, and desktop shells may cache an older icon. Windows and Linux
interactive icon checks remain required on those operating systems.
