<p align="center">
  <img src="Icon.png" alt="Net Test logo" width="150"/>
</p>

# Net Test

Net Test is a cross-platform desktop network diagnostics application built with
Go and [Wails](https://wails.io/). Its interface uses React and TypeScript and
runs inside the operating system's native webview.

The current application supports HTTP and HTTPS endpoint testing. TCP, ICMP,
and DNS tests are planned for future releases.

## Current status

The Wails application is under active development in `wails-http/`. The current
version is an HTTP/HTTPS desktop application rather than a complete release of
the planned network-testing suite.

Current functionality:

- Run multiple HTTP or HTTPS tests using GET, PUT, or PATCH.
- Select HTTP or HTTPS separately from the URL input. HTTPS is the default.
- Configure the probe interval and request timeout.
- Define successful responses using HTTP status groups or exact status codes.
- Connect directly or through an optional authenticated HTTP/HTTPS proxy.
- Verify target and proxy TLS certificates with the operating system trust
  store.
- Start, stop, rerun, search, filter, and permanently delete tests.
- Select any part of a test row to show that test's details and graph.
- View live response time, average latency, failure rate, and probe count.
- Inspect recent responses and open a graph in a separate desktop window.
- View the complete timeline from the test's start through its latest result.
- Zoom with two timeline handles and reset to the complete timeline.
- Pause graph updates while the test continues collecting and saving results.
- Save every completed probe automatically in a local SQLite database.
- Restore saved tests and complete timelines after restarting the application.
- Browse saved History with search, status filters, and pagination.
- Export every saved probe for a test to CSV.

## Planned functionality

The following test types will be added after the HTTP/HTTPS application is
stable across the supported desktop platforms:

- **TCP tests** for checking connections to a host and port.
- **DNS tests** with selectable UDP or TCP transport and custom DNS servers.
- **ICMP tests** with platform-specific permission and compatibility handling.

These protocols will reuse the current session lifecycle, automatic storage,
History, CSV export, and timeline graph components.

## Build from source

### Requirements

- Go 1.25 or later
- Node.js 24 LTS and npm
- Xcode Command Line Tools on macOS
- WebView2 on Windows
- GTK4 and WebKitGTK 6.0 development packages on Linux

The project currently pins Wails v3.0.0-beta.17. Wails v3 is still a beta, so
release builds should be tested on each target operating system.

### Fresh setup

Clone the repository and install both frontend and Go dependencies:

```bash
git clone https://github.com/djian01/nt_gui.git
cd nt_gui/wails-http
make setup
```

Build the frontend and desktop executable:

```bash
make build
```

The executable is created at:

```text
wails-http/bin/nt-http
```

Run the application with:

```bash
./bin/nt-http
```

You can also build and run it together:

```bash
make run
```

### Rebuild after Go-only changes

The React frontend is embedded in the executable. If `frontend/dist` already
exists and only Go code changed, compile directly with:

```bash
go build -tags production -o bin/nt-http .
```

A plain `go build .` may compile when `frontend/dist` exists, but it omits the
Wails production build tag and is not the recommended command for a
distributable executable.

### macOS application bundle

```bash
make mac-app
```

This creates an unsigned application bundle at:

```text
wails-http/bin/NT HTTP Prototype.app
```

### Windows executable

Run these commands from `wails-http` in PowerShell:

```powershell
npm --prefix frontend ci
npm --prefix frontend run build
go build -tags production -ldflags "-H windowsgui" -o bin/nt-http.exe .
```

### Linux executable

Install the GTK4 and WebKitGTK 6.0 development packages supplied by your Linux
distribution, then run:

```bash
make setup
make build
```

## Tests

Run the Go race tests and frontend type check/build from `wails-http`:

```bash
make test
```

The repository also contains a native CI build matrix for macOS, Windows, and
Ubuntu in `.github/workflows/wails-http.yml`.

## Data and privacy

Tests and probes are stored automatically in a fresh SQLite database under the
current user's configuration directory:

```text
nt-wails/results.db
```

There is no automatic probe-count retention limit. Results remain until the
user deletes the saved test. The graph loads indexed time ranges and compact
overview data so it does not need to load an entire long-running test into each
window.

Proxy passwords are used only for the running test. They are not returned to
the frontend, written to SQLite, or included in CSV exports. A password must be
entered again when rerunning a saved test that used authenticated proxy access.

## Project structure

```text
wails-http/
  main.go                    Wails application and desktop service
  Makefile                   Repeatable setup, build, run, and test commands
  internal/ping/
    runner.go                HTTP session lifecycle and network probes
    store.go                 SQLite persistence, timeline queries, and CSV data
  frontend/
    src/App.tsx              HTTP form, test table, History, and desktop actions
    src/LatencyChart.tsx     Live and saved timeline graph
    src/useSessions.ts       Session, event, and selected-test state
    src/api.ts               Typed Wails desktop API wrappers
    src/style.css            Dark blue application design
docs/
  architecture/              Architecture and implementation decisions
  api_reference/             Desktop method contracts and call flows
```

More detail is available in the
[architecture documentation](docs/architecture/http-wails-prototype.md) and
[desktop API reference](docs/api_reference/http-wails-prototype.md).

## License

This project is licensed under a **Personal-Use License**:

- Free for individuals for personal, educational, or hobby use.
- Organizational or commercial use, including businesses, institutions, and
  government, requires a commercial license.

For commercial licensing, contact
[dennis.jian@packetstreams.net](mailto:dennis.jian@packetstreams.net).

See [LICENSE](LICENSE) for the full terms and conditions.
