<p align="center">
  <img src="Icon.png" alt="Net Test logo" width="150"/>
</p>

# Net Test

Net Test is a cross-platform desktop application for running network tests,
tracking results, and investigating performance over time. It is built with Go
and [Wails](https://wails.io/); its React and TypeScript interface runs inside
the operating system's native webview.

## Current status

The desktop application is under active development in `desktop/`. Its shared
test lifecycle, persistence, History, CSV export, and timeline components form
the foundation for multiple network test types.

Current functionality:

- Run TCP connection tests for multiple hostnames/IPs and a port, with recording, charts, replay, and CSV analysis.
- Run DNS tests with custom resolver IPs and UDP/TCP transport.
- Run multiple HTTP or HTTPS tests using GET, POST, PUT, or PATCH, with optional redirect following.
- Import current or legacy HTTP, DNS, or TCP CSV results for analysis, export chart PNGs, and delete selected saved tests.
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
- Switch between persistent light and dark interface themes.

## Test types

HTTP/HTTPS, DNS, and TCP tests are available now. Additional test types planned
for future releases include:

- **ICMP tests** with platform-specific permission and compatibility handling.

These protocols will reuse the current session lifecycle, automatic storage,
History, CSV export, and timeline graph components.

## Build and install

See the [Build and installation guide](BUILD_AND_INSTALL.md) for numbered
macOS, Windows, and Linux setup, build, installation, and test instructions.

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
desktop/
  main.go                    Wails application and desktop service
  Makefile                   Repeatable setup, build, run, and test commands
  internal/testengine/
    runner.go                Shared HTTP/DNS/TCP session lifecycle and probes
    store.go                 SQLite persistence, timeline queries, and CSV data
  frontend/
    src/App.tsx              Protocol forms, test table, History, and desktop actions
    src/LatencyChart.tsx     Live and saved timeline graph
    src/useSessions.ts       Session, event, and selected-test state
    src/api.ts               Typed Wails desktop API wrappers
    src/style.css            Shared light and dark design standards
docs/
  architecture/              Architecture and implementation decisions
  api_reference/             Desktop method contracts and call flows
```

More detail is available in the
[architecture documentation](docs/architecture/desktop.md) and
[desktop API reference](docs/api_reference/desktop.md).

## License

This project is licensed under a **Personal-Use License**:

- Free for individuals for personal, educational, or hobby use.
- Organizational or commercial use, including businesses, institutions, and
  government, requires a commercial license.

For commercial licensing, contact
[dennis.jian@packetstreams.net](mailto:dennis.jian@packetstreams.net).

See [LICENSE](LICENSE) for the full terms and conditions.
