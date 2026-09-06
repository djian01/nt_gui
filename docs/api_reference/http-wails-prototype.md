# HTTP prototype desktop API

These APIs are Wails bound Go methods, **not HTTP endpoints**. Calls are local to the desktop runtime. `wails-http/frontend/src/api.ts` invokes `main.PingService.<Method>` using `Call.ByName`; `wails-http/main.go` registers `PingService` with `application.NewService`.

## Method contract

| Bound method | Request | Response | Consumer |
| --- | --- | --- | --- |
| `main.PingService.Start` | `Config` | `Session`, or validation/capacity error | `App.tsx: start`, `restart` |
| `main.PingService.List` | None | `Session[]`, oldest first; empty array when none | `useSessions.ts: useSessions` startup |
| `main.PingService.Get` | Session ID string | `Detail`, or not-found error | `useSessions.ts: useSessions` selection effect |
| `main.PingService.Stop` | Session ID string | Stopped `Session`, or not-found error | `App.tsx: stop`, main and chart windows |
| `main.PingService.Remove` | Stopped session ID string | Void, or active/not-found error | `App.tsx: App` row remove action |
| `main.PingService.OpenChart` | Session ID string | Void, or not-found error | `App.tsx: App` chart actions |

`Config` has `url` (complete HTTP/HTTPS URL), `method` (`GET`/`HEAD`), `intervalMs` (1000–60000), and `timeoutMs` (1000–30000). URLs are limited to 4096 characters and cannot contain credentials or fragments. Limits are 8 active and 24 retained sessions.

`Session` contains `id`, `config`, `running`, `startedAt`, nullable `endedAt`, monotonic `revision`, `sent`, `succeeded`, `minRtt`, `maxRtt`, `avgRtt`, and nullable `last`. Timestamps are RFC3339 strings; RTT fields are milliseconds. RTT aggregates include successful probes only. Empty aggregates are zero, rendered as a dash until there is a successful response. `sent` counts completed probes; explicit cancellation is excluded.

`Sample` contains `sequence` (one-based), `time`, `rtt`, `statusCode` (zero if no HTTP response), `success`, and `error` (empty on success). A received 200–399 response is successful; redirects are not followed. `Detail` contains `session` and `samples` (chronological, latest 600 at most).

## Events

| Event | Payload | Producer | Consumer |
| --- | --- | --- | --- |
| `http:updated` | `Session`, including latest sample | `runner.go: Runner.emitLocked` → `main.go: main` callback → `App.Event.Emit` | `api.ts: onUpdate` → `useSessions.ts: merge` → `App.tsx`, `LatencyChart.tsx` |
| `http:removed` | Session ID string | `main.go: PingService.Remove` → `App.Event.Emit` | `api.ts: onRemove` → `useSessions.ts: useSessions` |

Updates occur on start, completed probe, and stop. Each carries a summary and at most one sample, not the whole history. Consumers ignore stale revisions. Selection subscribes before fetching detail and reconciles updates received during the snapshot. Removed IDs are ignored if older queued updates arrive.

## Start and results flow

```mermaid
flowchart TD
    A["frontend/src/App.tsx: start or restart"] --> B["frontend/src/api.ts: api.start → Call.ByName"]
    B --> C["main.go: PingService.Start"]
    C --> D["internal/ping/runner.go: Runner.Start → validate"]
    D --> E["internal/ping/runner.go: Runner.run"]
    E --> F["internal/ping/runner.go: probe → http.Client.Do"]
    F --> G["internal/ping/runner.go: Runner.run updates counters and ring buffer"]
    G --> H["internal/ping/runner.go: Runner.emitLocked"]
    H --> I["main.go: main callback → App.Event.Emit http:updated"]
    I --> J["frontend/src/api.ts: onUpdate"]
    J --> K["frontend/src/useSessions.ts: merge"]
    K --> L["frontend/src/App.tsx: App / Metric; LatencyChart.tsx: LatencyChart"]
```

## Snapshot flow

```mermaid
flowchart TD
    A["frontend/src/useSessions.ts: useSessions startup"] --> B["frontend/src/api.ts: api.list"]
    B --> C["main.go: PingService.List"]
    C --> D["internal/ping/runner.go: Runner.List"]
    E["frontend/src/useSessions.ts: selection effect"] --> F["frontend/src/api.ts: api.get"]
    F --> G["main.go: PingService.Get"]
    G --> H["internal/ping/runner.go: Runner.Get copies bounded ring history"]
    D --> I["frontend/src/useSessions.ts: merge"]
    H --> J["frontend/src/useSessions.ts: reconcile pending events / setDetail"]
```

## Stop, remove, and shutdown flow

```mermaid
flowchart TD
    A["frontend/src/App.tsx: stop"] --> B["frontend/src/api.ts: api.stop"]
    B --> C["main.go: PingService.Stop"]
    C --> D["internal/ping/runner.go: Runner.Stop cancels request; Runner.emitLocked"]
    E["frontend/src/App.tsx: remove action"] --> F["frontend/src/api.ts: api.remove"]
    F --> G["main.go: PingService.Remove"]
    G --> H["internal/ping/runner.go: Runner.Remove checks stopped state"]
    H --> I["main.go: PingService.Remove → Window.Close / App.Event.Emit http:removed"]
    J["main.go: main WindowClosing callback → App.Quit"] --> K["main.go: OnShutdown → internal/ping/runner.go: Runner.Close"]
    K --> L["internal/ping/runner.go: cancel all contexts; WaitGroup.Wait"]
```

## Chart window flow

```mermaid
flowchart TD
    A["frontend/src/App.tsx: chart action"] --> B["frontend/src/api.ts: api.chart"]
    B --> C["main.go: PingService.OpenChart"]
    C --> D["internal/ping/runner.go: Runner.Get validates session"]
    D --> E["main.go: PingService.OpenChart → Window.GetByName"]
    E --> F["main.go: PingService.OpenChart → Window.Show / Focus if present"]
    E --> G["main.go: PingService.OpenChart → Window.NewWithOptions if absent"]
    G --> H["frontend/src/App.tsx: App reads chart query parameter"]
    H --> I["frontend/src/useSessions.ts: useSessions → api.get / onUpdate"]
    I --> J["frontend/src/LatencyChart.tsx: LatencyChart"]
```

The original Fyne application has no API changes. This prototype has no persistence or external control endpoint. Request errors omit the URL wrapper so query tokens are not duplicated into error messages; the user-entered URL itself remains part of session configuration and is displayed in the UI.
