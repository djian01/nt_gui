# Build and install Net Test

[Back to README](README.md)

This guide builds the current HTTP/HTTPS desktop application in `wails-http/`.
Run the instructions for your operating system on that operating system.
Commands after cloning assume your terminal is in `nt_gui/wails-http`.

Choose your platform:

- [macOS](#macos)
- [Windows](#windows)
- [Linux](#linux)
- [Rebuilding after changes](#rebuilding-after-changes)
- [Tests](#tests)
- [Troubleshooting](#troubleshooting)

## Requirements

All platforms need Git, Go **1.25 or later**, and **Node.js 24 LTS with npm**.
Use the installers or instructions from [Git](https://git-scm.com/downloads),
[Go](https://go.dev/doc/install), and [Node.js](https://nodejs.org/en/download).
Select downloads matching your operating system and processor architecture.
Dependency installation requires internet access.

The project pins Wails **v3.0.0-beta.17** in `wails-http/go.mod` and the matching
frontend runtime in `wails-http/frontend/package.json`. The commands below use
the repository's Makefile or Go directly; installing the Wails CLI separately is
not required. Wails v3 is still a beta, so release builds need testing on each
target operating system.

Native prerequisites are Xcode Command Line Tools on macOS, WebView2 on Windows,
and GTK4/WebKitGTK 6.0 development libraries on Linux. See the
[Wails platform requirements](https://v3.wails.io/quick-start/installation/)
for additional platform setup information.

These steps produce a local executable or an unsigned macOS application bundle.
The repository does not currently provide a signed installer workflow.

## macOS

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
   cd nt_gui/wails-http
   ```

   If you already have a checkout, open its `wails-http` directory instead.

4. **Install the project dependencies.**

   ```bash
   make setup
   ```

   This runs `npm --prefix frontend ci` using the frontend lockfile and
   `go mod download` using the Go module's dependency versions.

5. **Build and launch the executable.**

   ```bash
   make build
   ./bin/nt-http
   ```

   `make build` checks the TypeScript frontend, builds its assets, and embeds
   them in `wails-http/bin/nt-http`. Alternatively, `make run` builds and
   launches the application in one command. Quit the app before continuing.

6. **Create a macOS application bundle.**

   ```bash
   make mac-app
   ```

   This rebuilds the app and creates
   `wails-http/bin/NT HTTP Prototype.app`. The bundle uses the current prototype
   name. The Makefile sets matching macOS compiler and linker deployment targets
   of 11.0; that setting does not establish runtime compatibility with older
   macOS releases.

7. **Install and open the bundle for your user account.**

   ```bash
   mkdir -p "$HOME/Applications"
   ditto "bin/NT HTTP Prototype.app" "$HOME/Applications/NT HTTP Prototype.app"
   open "$HOME/Applications/NT HTTP Prototype.app"
   ```

   You can also open it from your home folder's `Applications` folder in Finder.
   Quit an existing installed copy before replacing it. The bundle is unsigned
   and is not notarized.

## Windows

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
   Set-Location nt_gui/wails-http
   ```

   If you already have a checkout, open its `wails-http` directory instead.

4. **Install the project dependencies.**

   ```powershell
   npm --prefix frontend ci
   go mod download
   ```

5. **Build the frontend and Windows executable.**

   ```powershell
   npm --prefix frontend run build
   go build -tags production -ldflags "-H windowsgui" -o bin/nt-http.exe .
   ```

   Run each command only after the previous command succeeds. The executable
   is created at `wails-http\bin\nt-http.exe`. The `windowsgui` linker option
   prevents a console window from opening alongside the desktop app.

6. **Launch the executable to check it opens.**

   ```powershell
   .\bin\nt-http.exe
   ```

   Quit the app before copying or rebuilding the executable.

7. **Install a copy for your user account.**

   ```powershell
   $installDir = Join-Path $env:LOCALAPPDATA "Programs\Net Test"
   New-Item -ItemType Directory -Force -Path $installDir | Out-Null
   Copy-Item .\bin\nt-http.exe -Destination $installDir -Force
   Start-Process (Join-Path $installDir "nt-http.exe")
   ```

   This copies the app to `%LOCALAPPDATA%\Programs\Net Test`; it does not create
   a Start menu entry or installer registration. You can create a shortcut to
   the executable in File Explorer. Quit an installed copy before replacing it.

## Linux

1. **Install Git and the native build dependencies.**

   On **Ubuntu 24.04**, matching the repository's CI configuration, run:

   ```bash
   sudo apt-get update
   sudo apt-get install -y git build-essential pkg-config libgtk-4-dev libwebkitgtk-6.0-dev
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
   cd nt_gui/wails-http
   ```

   If you already have a checkout, open its `wails-http` directory instead.

4. **Install the project dependencies.**

   ```bash
   make setup
   ```

   This installs the locked frontend dependencies and downloads Go modules.

5. **Build and launch the executable.**

   ```bash
   make build
   ./bin/nt-http
   ```

   The executable is created at `wails-http/bin/nt-http`. Use a terminal in a
   graphical desktop session so the application can open its window.
   Alternatively, `make run` builds and launches the app together.

6. **Install the executable for your user account.**

   Quit the app, then run:

   ```bash
   mkdir -p "$HOME/.local/bin"
   install -m 755 bin/nt-http "$HOME/.local/bin/nt-http"
   "$HOME/.local/bin/nt-http"
   ```

   If `$HOME/.local/bin` is on your `PATH`, you can launch it with `nt-http`.
   This does not create a desktop menu entry. GTK4 and WebKitGTK 6.0 runtime
   libraries must remain installed; copying this executable to another Linux
   machine does not install its system libraries.

## Rebuilding after changes

1. **Enter `wails-http` and close the running app.** If dependency manifests or
   lockfiles changed, rerun `make setup` on macOS/Linux, or
   `npm --prefix frontend ci` followed by `go mod download` on Windows.

2. **Choose the appropriate build.** For frontend changes or a complete rebuild,
   use `make build` on macOS/Linux. On Windows, repeat the frontend and Go build
   commands in Windows step 5.

   For **Go-only changes**, if `frontend/dist` already exists and is current,
   you can skip the frontend build:

   macOS/Linux:

   ```bash
   go build -tags production -o bin/nt-http .
   ```

   On macOS, prefer `make build` when you need the Makefile's explicit deployment
   target settings as well.

   Windows PowerShell:

   ```powershell
   go build -tags production -ldflags "-H windowsgui" -o bin/nt-http.exe .
   ```

   A plain `go build .` may compile when frontend assets exist, but omits the
   Wails `production` build tag and is not the recommended distributable build.

3. **Update the installed copy.** Repeat your platform's installation step.
   On macOS, run `make mac-app` first to refresh the bundle. Rebuilding alone
   does not replace a previously installed copy outside the repository.

## Tests

1. **Install the dependencies** using your platform's setup steps and enter
   `wails-http`.

2. **Run the Go race tests and frontend type check/build.**

   macOS/Linux:

   ```bash
   make test
   ```

   Windows PowerShell:

   ```powershell
   go test -race ./internal/ping
   npm --prefix frontend run build
   ```

   Go's race detector requires CGO and a compatible C compiler. On Windows,
   install a Go-compatible MinGW-w64 compiler and make `gcc` available on
   `PATH` if you want to run race tests. See the
   [Go race detector requirements](https://go.dev/doc/articles/race_detector#Requirements).

3. **Build and launch the desktop app** to check the native UI on the target OS.
   Automated tests and the frontend build do not replace this check.

The [native CI workflow](.github/workflows/wails-http.yml) defines builds and
race tests for macOS, Windows, and Ubuntu 24.04. Its presence does not establish
that a particular revision has passed on those platforms; check its run results.

## Troubleshooting

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
