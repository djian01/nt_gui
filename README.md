<p align="center"><img src="Icon.png" alt="NET-Test logo" width="150"/></p>

# NET-Test v2.0.0

NET-Test is a network diagnostics desktop application built with Go, Wails, React, and TypeScript. The sole application is in `desktop/`.

## Features

- HTTP/HTTPS probes with GET, POST, PUT, PATCH, status validation, redirect control, and authenticated proxy support.
- DNS queries over UDP/TCP, TCP connection tests, and IPv4 ICMP echo tests with multiple targets.
- Start, stop, replay, recording, searchable paginated History, and confirmed deletion.
- Live and saved timelines, range selection, pause/reset, detached chart windows, and PNG export.
- Current and legacy HTTP/DNS/TCP/ICMP CSV import and recorded-probe CSV export.
- Persistent light/dark themes and distinct hover and selected row states.

Recording defaults to off: temporary tests remain in memory until closed or the app exits. Turning recording on saves future probes only. Saved results survive restart; tests never restart automatically.

## Build

From the repository root:

```sh
make setup
make build
make run
make test
# macOS bundle:
make mac-app
```

See [build and installation](BUILD_AND_INSTALL.md) for platform requirements. Go commands run inside `desktop/`; the root Makefile forwards to it.

## Data and compatibility

Existing desktop data stays in the user configuration directory at `nt-wails/results.db`. The application identity and saved theme preference are retained across the rename. Legacy `ntdata.db` databases are not migrated; legacy CSV exports can be imported as independent stopped sessions. No existing user database is deleted by this source migration.

TLS verification is enabled for targets and proxies. Proxy passwords are never saved in SQLite or exported. Ten tests may run concurrently. Recorded timelines use indexed summaries; unrecorded long runs can consume growing memory. Saved data remains until explicit deletion.

## Source and documentation

- `desktop/main.go`: Wails startup, local service bridge, dialogs, and chart windows.
- `desktop/internal/testengine/`: protocol adapters, lifecycle, recording, SQLite, and CSV formats.
- `desktop/frontend/src/`: forms, session state, charts, and themes.
- [Architecture](docs/architecture/desktop.md), [API contracts and flows](docs/api_reference/desktop.md), and [branding](docs/architecture/branding.md).

## License

[Personal-Use License](LICENSE): free for private, individual, non-commercial use, including personal learning and hobbies. Commercial and organizational use—including freelance work, institutional education, and nonprofit use—requires a separate written commercial license. Contact [dennis.jian@packetstreams.net](mailto:dennis.jian@packetstreams.net).

NET-Test is source-available under this restricted-use license. Third-party components retain their own licenses; see [third-party notices](THIRD_PARTY_NOTICES.txt).
