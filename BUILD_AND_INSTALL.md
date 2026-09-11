# Build and install NET-Test v2.0.1

[Back to README](README.md)

This guide builds the NET-Test desktop application in `desktop/`.
Run the instructions for your operating system on that operating system.
Choose one of the two workflows:

1. [Use scripts to build installation packages](#1-use-scripts-to-build-installation-packages).
2. [Use commands to build and install the executable](#2-use-commands-to-build-and-install-the-executable), including prerequisite tool installation.

Packaging scripts run from the repository root. Manual build and install
commands run from `nt_gui/desktop`, after installing the prerequisites and
cloning the repository.

## 1. Use scripts to build installation packages

Use the script matching your build machine's OS. Scripts can also be called
by absolute path from another directory. They install locked project dependencies,
build the frontend and native app, and create a package in the root
`installation_package/` folder. Users receiving a package do not need Go, Node.js,
Git, or the source code.

Everything in `installation_package/` is ignored by Git except `.gitkeep`.
Scripts and packaging configuration live in `scripts/` and are versioned.
Temporary packaging files are removed automatically. An existing package with
the same version and architecture is replaced only after packaging succeeds.

Install the prerequisites using your OS's commands below, then run the packaging
commands from the repository root. If you do not have the source yet, run
`git clone https://github.com/djian01/nt_gui.git` and `cd nt_gui` after installing
Git. Run each command only after the previous command succeeds. Use native builds:
do not override `GOOS` or `GOARCH`.
The default version comes from `desktop/build/darwin/Info.plist` (currently 2.0.1).
An optional version must use `major.minor.patch`; it labels the package, and on
macOS also updates the packaged bundle's version without editing the source plist.

### macOS DMG

Requires macOS, Xcode Command Line Tools (including Swift), Go, Node.js/npm,
Make, Python 3.10 or later, and `dmgbuild` 1.6.7. `hdiutil` is included with macOS.
Install Apple's tools first and finish the installation dialog before continuing:

```bash
xcode-select --install
```

If already installed, skip that command. Install [Homebrew](https://brew.sh/) if
`brew` is unavailable:

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

Follow the installer's printed **Next steps** to add Homebrew to your shell.
Then install Git, Go, [Node.js 24](https://formulae.brew.sh/formula/node@24), and Python:

```bash
brew install git go node@24 python@3.13
export PATH="$(brew --prefix node@24)/bin:$PATH"
git --version
go version
node --version
npm --version
make --version
swift --version
python3.13 --version
```

Confirm Go is at least 1.25 and Node.js reports `v24.x`. Repeat the Node.js
`export PATH` command in each new Terminal session, or add it to `~/.zshrc`.
Apple's tools provide Make and Swift.

Install the layout tool in a local, Git-ignored environment once, from the
repository root:

```bash
python3.13 -m venv .local/packaging-venv
.local/packaging-venv/bin/python -m pip install 'dmgbuild==1.6.7'
```

The explicit `python3.13` command avoids Apple's older system Python.
An existing environment created with an older Python must be
recreated with the newer interpreter. dmgbuild versions below 1.6.7 are rejected
because their backgrounds may appear blank on macOS 26.2 and later.

Then build from the repository root (the script also finds this local tool
automatically, or uses `dmgbuild` from `PATH`):

```bash
bash scripts/package-macos.sh
# Optional explicit package version:
bash scripts/package-macos.sh 2.0.1
```

Output: `installation_package/NET-Test-2.0.1-macos-arm64.dmg` on Apple Silicon,
or `...-amd64.dmg` on an Intel build machine. These are separate native packages,
not a universal binary.

Users open the DMG and drag **NET-Test** onto **Applications**. Quit an installed
copy before replacing it. The window opens in icon view with NET-Test on the
left, a right-pointing arrow, and Applications on the right, plus installation
instructions below. A blue package-box icon distinguishes the DMG and mounted
volume from the app itself. The volume is named **NET-Test Installer**.
Layout metadata is written without automating Finder.

The DMG file's custom Finder icon uses macOS extended metadata, which some upload,
download, or copy tools strip. The mounted volume icon and installation layout
are embedded inside the DMG and survive such transfers.

The script cleans extended attributes on the staged app, ad-hoc signs the complete
bundle after final version metadata is written, and verifies its sealed resources.
It verifies the DMG and mounts it read-only to verify the bundled app again before
keeping the installer. An invalid bundle signature fails packaging and CI.

Ad-hoc signing provides integrity checks, not Apple developer trust. The app is
not Developer ID signed or notarized; downloaded copies may still need explicit
approval in System Settings > Privacy & Security. Developer ID signing and Apple
notarization are separate release steps and are not automated by these scripts.

### Windows installation wizard

Requires Windows with native amd64 or arm64 Go, Node.js/npm, and
[NSIS 3](https://nsis.sourceforge.io/Download). In PowerShell, install the tools
with WinGet (provided by Microsoft's **App Installer**):

```powershell
winget install --exact --id Git.Git
winget install --exact --id GoLang.Go
winget install --exact --id OpenJS.NodeJS.LTS
winget install --exact --id NSIS.NSIS
winget install --exact --id Microsoft.EdgeWebView2Runtime
```

Close and reopen PowerShell to refresh `PATH`. Add the NSIS installation folder
for this session (adjust if you chose a different installation directory):

```powershell
$env:Path = "${env:ProgramFiles(x86)}\NSIS;$env:Path"
git --version
go version
node --version
npm.cmd --version
makensis /VERSION
```

Confirm Go is at least 1.25 and Node.js reports `v24.x`. The LTS WinGet package
follows the active LTS major; if it supplies a different major, install Node.js
24 from the [official downloads](https://nodejs.org/en/download) instead.
WebView2 is needed to launch the app. Make and a separate Wails CLI are not needed.
Repeat the NSIS `PATH` command in new sessions, or add its directory through
Windows **Environment Variables**. Then, from PowerShell at the repository root:

```powershell
.\scripts\package-windows.ps1
# Optional explicit package version:
.\scripts\package-windows.ps1 -Version 2.0.1
```

If local script execution is blocked, run the reviewed script with a policy
override limited to this process:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\package-windows.ps1
```

Output: `installation_package\NET-Test-2.0.1-windows-amd64-Setup.exe`
(or `arm64`). Users run the wizard to install under
`%LOCALAPPDATA%\Programs\net-test`. It creates a Start menu shortcut and an
Installed Apps uninstall entry for the current user, without requiring admin.
Quit NET-Test before upgrading or uninstalling. Uninstall keeps saved user data.
Setup uses the distinct package-box icon; the installed app and its shortcut
continue to use the NET-Test app icon.

The installer checks for Microsoft's registered WebView2 Evergreen Runtime.
If absent, it stops with instructions to install the runtime from the
[Microsoft download page](https://developer.microsoft.com/microsoft-edge/webview2/)
and rerun Setup. WebView2 is not bundled. The generated installer is unsigned;
Windows may show an unknown-publisher or SmartScreen warning. Code signing is
a separate release step.

### Linux DEB / RPM

Install Git, Go, and the native libraries using the commands for your distribution.
For **Ubuntu 24.04 / DEB**, the Go backports PPA provides newer Go versions than
the default repository (see the [Go Ubuntu guide](https://go.dev/wiki/Ubuntu)):

```bash
sudo apt-get update
sudo apt-get install -y git curl ca-certificates build-essential pkg-config libgtk-4-dev libwebkitgtk-6.0-dev desktop-file-utils software-properties-common
sudo add-apt-repository -y ppa:longsleep/golang-backports
sudo apt-get update
sudo apt-get install -y golang-go
```

For **Fedora / RPM**:

```bash
sudo dnf install -y git curl ca-certificates golang gcc gcc-c++ make pkgconf-pkg-config gtk4-devel webkitgtk6.0-devel desktop-file-utils
```

On either distribution, install Node.js 24 with [nvm](https://github.com/nvm-sh/nvm)
in Bash. If nvm is already installed, load it and run the final two commands:

```bash
curl -fsSL https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.3/install.sh | bash
export NVM_DIR="$HOME/.nvm"
. "$NVM_DIR/nvm.sh"
nvm install 24
nvm alias default 24
```

Check the prerequisites before proceeding:

```bash
git --version
go version
node --version
npm --version
make --version
pkg-config --modversion gtk4 webkitgtk-6.0
```

Go must be at least 1.25, Node.js must report `v24.x`, and the last command must
report both libraries. If your distribution supplies older Go, use the
[official Go installer instructions](https://go.dev/doc/install) before continuing.
The Ubuntu PPA commands are Ubuntu-specific; do not run them on Debian.

Then install [nFPM](https://nfpm.goreleaser.com/docs/install/) and make its executable
available in the same shell:

```bash
go install github.com/goreleaser/nfpm/v2/cmd/nfpm@v2.47.0
export PATH="$(go env GOPATH)/bin:$PATH"
nfpm --version
```

Build DEB packages on Ubuntu 24.04 or a compatible Debian-family target.
Build RPM packages on the Fedora-family target you intend to support, with
`gtk4-devel`, `webkitgtk6.0-devel`, a C/C++ toolchain, Make, and `pkg-config`.
From the repository root:

```bash
bash scripts/package-linux.sh deb
bash scripts/package-linux.sh rpm
# Optional explicit package version:
bash scripts/package-linux.sh deb 2.0.1
```

Outputs: `installation_package/NET-Test-2.0.1-linux-amd64.deb` or `.rpm`
(architecture follows native Go). Users install a downloaded package with:

```bash
# Ubuntu / Debian, from the folder containing the package:
sudo apt install ./NET-Test-2.0.1-linux-amd64.deb
# Fedora:
sudo dnf install ./NET-Test-2.0.1-linux-amd64.rpm
```

Packages install the executable, application menu entry, icons, license, and
third-party notices.
They declare GTK4, WebKitGTK 6.0, and CA certificate dependencies, which the
package manager resolves from configured repositories. Use `sudo apt remove
net-test` or `sudo dnf remove net-test` to uninstall; user data is retained.

Linux packages are not portable across all distributions or releases: native
library versions and package names must match. Building an RPM on Ubuntu does
not establish Fedora compatibility. Test installation and launch on each target
distribution and architecture before distributing it.

### Packaging help and verification

Release license files are `LICENSE` and `THIRD_PARTY_NOTICES.txt`. The macOS
bundle stores them in `Contents/Resources/`; Linux packages and archives use
`/usr/share/licenses/net-test/`; the Windows installer places them next to the
executable and removes them on uninstall. Windows CI artifacts also include
both files alongside the executable.

When runtime dependencies change, refresh `THIRD_PARTY_NOTICES.txt` from the
versions in `desktop/go.mod` and production entries in
`desktop/frontend/package-lock.json`. Preserve upstream license and copyright
text, including nested third-party notices. The Wails runtime npm package omits
its license file; use the matching Wails source module's MIT license. Include
the Go runtime license, and recheck any newly bundled platform components.
The current inventory excludes development-only tools and separately installed
system frameworks. It does not establish exhaustive legal compliance.

Installer artwork lives in `scripts/installer-assets/`: edit `installer.svg` or
`dmg-background.svg`, then run `make icons` after dependency setup to regenerate
the PNG/ICO/ICNS assets. Commit these assets alongside their SVG sources. Finder
icon positions and window size are in `scripts/dmg-settings.py`; keep them aligned
with the background. `scripts/set-macos-file-icon.swift` applies the DMG file icon.

```bash
bash scripts/package-macos.sh --help
bash scripts/package-linux.sh --help
```

On Windows use `.\scripts\package-windows.ps1 -Help`. Packaging scripts stop on
build errors; they do not run the application test suite or install the output.
Run the existing tests and verify install, launch, upgrade, and uninstall on each
target OS before a release. Signing, notarization, automatic updates, and CI
release publishing are not included.

## 2. Use commands to build and install the executable

Follow your platform's steps in order: install prerequisite tools, get the
source, install project dependencies, build the executable, and install a local
copy using commands. These steps do not run the installation-package scripts.

Choose your platform:

- [macOS](#macos)
- [Windows](#windows)
- [Linux](#linux)
- [Rebuilding after changes](#rebuilding-after-changes)
- [Tests](#tests)
- [Troubleshooting](#troubleshooting)

### Which command should I use?

Run all commands from `nt_gui/desktop`.

| Command | When to use it | Result |
| --- | --- | --- |
| `make setup` | First setup, or after `package.json`, `package-lock.json`, `go.mod`, or `go.sum` changes | Installs the locked frontend and Go dependencies |
| `make run` | Normal development | Rebuilds the frontend and executable, then launches `bin/net-test` |
| `make build` | You need a rebuilt executable without launching it | Rebuilds `bin/net-test`; it does not refresh a macOS `.app` bundle |
| `make mac-app` | You want to open or distribute the macOS app bundle | Rebuilds everything and refreshes `bin/NET-Test.app` |
| `make linux-package` | You want the Linux executable, launcher, and icons together | Creates `bin/net-test-linux/usr/` with the executable and desktop integration files |
| `make icons` | You changed the logo vector | Regenerates the committed PNG, Windows ICO, and macOS ICNS assets |

You do not need to run `make setup` before every build. Both `make run` and
`make mac-app` invoke the required build steps automatically. If you previously
copied the app outside the repository, rebuild the bundle and copy it again;
none of the build commands replace an installed copy automatically.

### Requirements

All platforms need Git, Go **1.25 or later**, and **Node.js 24 LTS with npm**.
Use the installers or instructions from [Git](https://git-scm.com/downloads),
[Go](https://go.dev/doc/install), and [Node.js](https://nodejs.org/en/download).
Select downloads matching your operating system and processor architecture.
Dependency installation requires internet access.

The project pins Wails **v3.0.0-beta.17** in `desktop/go.mod` and the matching
frontend runtime in `desktop/frontend/package.json`. The commands below use
the repository's Makefile or Go directly; installing the Wails CLI separately is
not required. Wails v3 is still a beta, so release builds need testing on each
target operating system.

Native prerequisites are Xcode Command Line Tools on macOS, WebView2 on Windows,
and GTK4/WebKitGTK 6.0 development libraries on Linux. See the
[Wails platform requirements](https://v3.wails.io/quick-start/installation/)
for additional platform setup information.

The manual steps below produce a local executable, Linux desktop files, or an
unsigned macOS application bundle. Use the packaging scripts above for installers.
The repository does not currently provide a signed installer workflow.

### macOS

1. **Install the development tools.**

   Install Go and Node.js using the links above, then install Apple's command
   line tools, which provide the native compiler, Git, and Make:

   ```bash
   xcode-select --install
   ```

   Finish the installation dialog before continuing. If the tools are already
   installed, continue to the next step. Open a new Terminal window after
   installing Go and Node.js so their commands are available.

2. **Check the tools.**

   ```bash
   git --version
   go version
   node --version
   npm --version
   make --version
   xcode-select -p
   ```

   Confirm Go is at least 1.25 and Node.js reports `v24.x`.

3. **Clone the repository and enter the application directory.**

   ```bash
   git clone https://github.com/djian01/nt_gui.git
   cd nt_gui/desktop
   ```

   If you already have a checkout, open its `desktop` directory instead.

4. **Install the project dependencies.**

   ```bash
   make setup
   ```

   This runs `npm --prefix frontend ci` using the frontend lockfile and
   `go mod download` using the Go module's dependency versions.

5. **Build and launch during development.**

   ```bash
   make build
   ./bin/net-test
   ```

   `make build` checks and builds the TypeScript frontend, embeds its assets in
   `desktop/bin/net-test`, and `./bin/net-test` launches it. You can also use `make run` to build and
   launch in one command. Quit the app before continuing.

6. **Create a macOS application bundle.**

   ```bash
   make mac-app
   ```

   This rebuilds the app and creates
   `desktop/bin/NET-Test.app`. The Makefile sets matching macOS compiler and
   linker deployment targets of 11.0; that setting does not establish runtime
   compatibility with older macOS releases.

   The bundle includes `Contents/Resources/net-test.icns`, referenced by
   `CFBundleIconFile` in `Info.plist`, for Finder and Dock branding.

7. **Install and open the bundle for your user account.**

   ```bash
   mkdir -p "$HOME/Applications"
   ditto "bin/NET-Test.app" "$HOME/Applications/NET-Test.app"
   open "$HOME/Applications/NET-Test.app"
   ```

   You can also open it from your home folder's `Applications` folder in Finder.
   Quit an existing installed copy before replacing it. The bundle is unsigned
   and is not notarized.

### Windows

1. **Install the prerequisites.**

   Install Git for Windows, Go 1.25 or later, and Node.js 24 LTS with npm using
   the links above. Ensure their installers add the commands to `PATH`.
   Install the [Microsoft Edge WebView2 Runtime](https://developer.microsoft.com/en-us/microsoft-edge/webview2/)
   if it is not already available. WebView2 is needed to display the desktop UI.

   Open a new **PowerShell** window after installation. The commands below do
   not require Make.

2. **Check the tools.**

   ```powershell
   git --version
   go version
   node --version
   npm --version
   ```

   Confirm Go is at least 1.25 and Node.js reports `v24.x`. If PowerShell blocks
   `npm.ps1`, use `npm.cmd` in place of `npm` in this guide.

3. **Clone the repository and enter the application directory.**

   ```powershell
   git clone https://github.com/djian01/nt_gui.git
   Set-Location nt_gui/desktop
   ```

   If you already have a checkout, open its `desktop` directory instead.

4. **Install the project dependencies.**

   ```powershell
   npm --prefix frontend ci
   go mod download
   ```

5. **Build the frontend and Windows executable.**

   ```powershell
   npm --prefix frontend run build
   $targetArch = go env GOARCH
   go run github.com/akavel/rsrc@v0.10.2 -arch $targetArch -ico build/icons/net-test.ico -o "rsrc_windows_$targetArch.syso"
   go build -tags production -ldflags "-H windowsgui" -o bin/net-test.exe .
   ```

   Run each command only after the previous command succeeds. The executable
   is created at `desktop\bin\net-test.exe`. The `windowsgui` linker option
   prevents a console window from opening alongside the desktop app.

   The resource step embeds the multi-size icon into the executable for Explorer
   and shortcuts. It uses the pinned `rsrc` tool, downloaded on first use, and
   must succeed before building. The architecture-specific `.syso` stays local
   and is ignored by Git. Windows `make build` runs this step automatically.

6. **Launch the executable to check it opens.**

   ```powershell
   .\bin\net-test.exe
   ```

   Quit the app before copying or rebuilding the executable.

7. **Install a copy for your user account.**

   ```powershell
   $installDir = Join-Path $env:LOCALAPPDATA "Programs\net-test"
   New-Item -ItemType Directory -Force -Path $installDir | Out-Null
   Copy-Item .\bin\net-test.exe -Destination $installDir -Force
   Start-Process (Join-Path $installDir "net-test.exe")
   ```

   This copies the app to `%LOCALAPPDATA%\Programs\net-test`; it does not create
   a Start menu entry or installer registration. You can create a shortcut to
   the executable in File Explorer. Quit an installed copy before replacing it.

### Linux

1. **Install Git and the native build dependencies.**

   On **Ubuntu 24.04**, matching the repository's CI configuration, run:

   ```bash
   sudo apt-get update
   sudo apt-get install -y git build-essential pkg-config libgtk-4-dev libwebkitgtk-6.0-dev desktop-file-utils
   ```

   For other distributions, install equivalent packages providing a C/C++
   compiler, Make, `pkg-config`, GTK4, and WebKitGTK 6.0 development files.
   Check availability for your distribution release; the commands in this
   guide use the project's default GTK4 build.

2. **Install and check Go and Node.js.**

   Follow the official Go and Node.js instructions linked above to install Go
   1.25 or later and Node.js 24 LTS with npm. Distribution repositories may
   provide older versions, so verify before continuing:

   ```bash
   git --version
   go version
   node --version
   npm --version
   make --version
   pkg-config --modversion gtk4 webkitgtk-6.0
   ```

   The last command should report versions for both native libraries.

3. **Clone the repository and enter the application directory.**

   ```bash
   git clone https://github.com/djian01/nt_gui.git
   cd nt_gui/desktop
   ```

   If you already have a checkout, open its `desktop` directory instead.

4. **Install the project dependencies.**

   ```bash
   make setup
   ```

   This installs the locked frontend dependencies and downloads Go modules.

5. **Build and launch the executable.**

   ```bash
   make build
   ./bin/net-test
   ```

   The executable is created at `desktop/bin/net-test`. Use a terminal in a
   graphical desktop session so the application can open its window.
   Alternatively, `make run` builds and launches the app together.

6. **Install the executable and launcher for your user account.**

   Quit the app, then run:

   ```bash
   mkdir -p "$HOME/.local/bin"
   install -m 755 bin/net-test "$HOME/.local/bin/net-test"
   appDataDir="${XDG_DATA_HOME:-$HOME/.local/share}"
   mkdir -p "$appDataDir/icons/hicolor/scalable/apps" "$appDataDir/icons/hicolor/512x512/apps"
   install -m 644 frontend/public/net-test.svg "$appDataDir/icons/hicolor/scalable/apps/net-test.svg"
   install -m 644 build/icons/net-test.png "$appDataDir/icons/hicolor/512x512/apps/net-test.png"
   desktop-file-install --dir="$appDataDir/applications" \
     --set-key=Exec --set-value="\"$HOME/.local/bin/net-test\"" \
     --set-key=TryExec --set-value="$HOME/.local/bin/net-test" \
     build/linux/net-test.desktop
   if command -v gtk-update-icon-cache >/dev/null 2>&1; then
     gtk-update-icon-cache -f -t "$appDataDir/icons/hicolor"
   fi
   "$HOME/.local/bin/net-test"
   ```

   If `$HOME/.local/bin` is on your `PATH`, you can launch it with `net-test`.
   The NET-Test menu entry uses the new icon and an absolute executable path,
   so it also works when the desktop session does not include that directory
   in `PATH`. Linux uses a launcher and theme icons rather than an icon embedded
   in the ELF executable. GTK4 and WebKitGTK 6.0 runtime
   libraries must remain installed; copying this executable to another Linux
   machine does not install its system libraries.

   For distribution, `make linux-package` collects the executable, desktop
   entry, SVG, and PNG under `bin/net-test-linux/usr/`. Install this tree under
   `/usr` using your distribution's packaging tools. The CI archive preserves
   this layout and executable permissions; it is not an automatic installer.

### Rebuilding after changes

1. **Enter `desktop` and close the running app.** You do not need `make setup`
   for ordinary source or styling changes. Run it again only when
   `package.json`, `package-lock.json`, `go.mod`, or `go.sum` changes. On
   Windows, use `npm --prefix frontend ci` followed by `go mod download` for
   the equivalent dependency refresh.

2. **Choose how you will launch the result.** For normal development on macOS
   or Linux, use `make run`; it rebuilds before launching. Use `make build` when
   you need only the refreshed `bin/net-test` executable. On Windows, repeat
   the frontend and Go build commands in Windows step 5.

   To refresh and open the macOS bundle inside the repository:

   ```bash
   make mac-app
   open "bin/NET-Test.app"
   ```

   Running `make build` does not update `bin/NET-Test.app` because the bundle
   contains a separate copied executable.

   For **Go-only changes**, if `frontend/dist` already exists and is current,
   you can skip the frontend build:

   macOS/Linux:

   ```bash
   go build -tags production -o bin/net-test .
   ```

   On macOS, prefer `make build` when you need the Makefile's explicit deployment
   target settings as well.

   Windows PowerShell:

   ```powershell
   $targetArch = go env GOARCH
   go run github.com/akavel/rsrc@v0.10.2 -arch $targetArch -ico build/icons/net-test.ico -o "rsrc_windows_$targetArch.syso"
   go build -tags production -ldflags "-H windowsgui" -o bin/net-test.exe .
   ```

   A plain `go build .` may compile when frontend assets exist, but omits the
   Wails `production` build tag and is not the recommended distributable build.

3. **Update an installed copy when needed.** Repeat your platform's installation
   step. On macOS, run `make mac-app` first, quit the installed app, and copy the
   refreshed bundle again. Building or refreshing the repository copy does not
   replace an app in `$HOME/Applications`, `/Applications`, or another folder.

### Updating the logo

Edit `desktop/frontend/public/net-test.svg`, then run `make icons` from
`desktop/` after `make setup`. The generator uses the pinned Sharp development
dependency and writes `Icon.png` plus `desktop/build/icons/net-test.png`,
`net-test.ico`, and `net-test.icns`. Commit the SVG and generated icon assets
together. Normal builds consume these assets without regenerating them.
Rebuild each native package after regeneration, including the Windows resource
step above. See [branding architecture](docs/architecture/branding.md).

### Tests

1. **Install the dependencies** using your platform's setup steps and enter
   `desktop`.

2. **Run the Go race tests and frontend type check/build.**

   macOS/Linux:

   ```bash
   make test
   ```

   Windows PowerShell:

   ```powershell
   go test -race ./internal/testengine
   npm --prefix frontend run build
   ```

   Go's race detector requires CGO and a compatible C compiler. On Windows,
   install a Go-compatible MinGW-w64 compiler and make `gcc` available on
   `PATH` if you want to run race tests. See the
   [Go race detector requirements](https://go.dev/doc/articles/race_detector#Requirements).

3. **Build and launch the desktop app** to check the native UI on the target OS.
   Automated tests and the frontend build do not replace this check.

The [native CI workflow](.github/workflows/desktop.yml) defines builds and
race tests for macOS, Windows, and Ubuntu 24.04. Its presence does not establish
that a particular revision has passed on those platforms; check its run results.

### Troubleshooting

| Symptom | Action |
| --- | --- |
| `go`, `node`, `npm`, or `git` is not found | Install the missing tool, reopen the terminal, and verify its version and `PATH`. |
| Go reports no files matching `frontend/dist` | Run the frontend build before compiling Go; the app embeds that directory. |
| npm reports an unsupported Node.js engine | Check that the active Node.js version is 24 LTS. |
| `npm ci` reports a lockfile mismatch | Use matching committed `package.json` and `package-lock.json` files; do not delete the lockfile as a setup workaround. |
| Linux cannot find `gtk4` or `webkitgtk-6.0` | Install the development packages and recheck with `pkg-config --modversion gtk4 webkitgtk-6.0`. |
| Linux cannot open a display | Launch from a graphical desktop session with access to its display. |
| Windows cannot initialize WebView2 | Install or repair the Microsoft Edge WebView2 Runtime and retry. |
| A rebuilt app still shows the old UI | Rebuild the frontend and executable, then replace the installed copy; on macOS recreate the bundle too. |
