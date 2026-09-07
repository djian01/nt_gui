# HTTP prototype desktop API

These APIs are Wails bound Go methods, **not HTTP endpoints**. Calls are local to the desktop runtime. `wails-http/frontend/src/api.ts` invokes `main.PingService.<Method>` using `Call.ByName`; `wails-http/main.go` registers `PingService` with `application.NewService`.

## Method contract

| Bound method | Request | Response | Consumer |
| --- | --- | --- | --- |
| `main.PingService.Start` | `Config` | `Session`, or validation/capacity error | `App.tsx: start` |
| `main.PingService.List` | None | `Session[]`, oldest first; empty array when none | `useSessions.ts: useSessions` startup |
| `main.PingService.Get` | Session ID string | `Detail`, or not-found error | `useSessions.ts: useSessions` selection effect |
| `main.PingService.Stop` | Session ID string | Stopped `Session`, or not-found error | `App.tsx: stop`, main and chart windows |
| `main.PingService.Restart` | Stopped session ID string | New running `Session`, or active/not-found/capacity error | `App.tsx: restart`, main and chart windows |
| `main.PingService.Remove` | Stopped session ID string | Void, or active/not-found error | `App.tsx: App` row remove action |
| `main.PingService.OpenChart` | Session ID string | Void, or not-found error | `App.tsx: App` chart actions |

`Config` has `url` (complete HTTP/HTTPS URL), `method` (`GET`/`PUT`/`PATCH`), `intervalMs` (1000–60000), `timeoutMs` (1000–30000), `acceptedStatuses`, and `proxy`. The form stores the address without a scheme and combines it with the selected HTTP/HTTPS value before `api.start`; HTTPS is the interface default. `acceptedStatuses` contains one to 20 unique values: status groups `2xx` through `5xx`, exact codes from 200 through 599, or a combination. `proxy` contains `enabled`, an HTTP/HTTPS proxy `url`, and optional `username`/`password`. Proxy URLs cannot contain credentials, a path, query, or fragment; credentials are supplied through their separate fields. URLs are limited to 4096 characters and cannot contain credentials or fragments. Limits are 8 active and 24 retained sessions.

`Session` contains `id`, `config`, `running`, `startedAt`, nullable `endedAt`, monotonic `revision`, `sent`, `succeeded`, `minRtt`, `maxRtt`, `avgRtt`, and nullable `last`. Timestamps are RFC3339 strings; RTT fields are milliseconds. RTT aggregates include successful probes only. Empty aggregates are zero, rendered as a dash until there is a successful response. `sent` counts completed probes; explicit cancellation is excluded.

`Sample` contains `sequence` (one-based), `time`, `rtt`, `statusCode` (zero if no HTTP response), `success`, and `error` (empty on success). A received response is successful when it matches any configured status group or exact code; redirects are not followed. `Detail` contains `session` and every sample collected by that session in chronological order. History is held in memory until the session is removed or the app quits.

Returned session configuration never contains a proxy password. The runner retains the credential privately for the session lifetime so `Restart` can create a new session with the same configuration. Restarting does not mutate or remove the stopped source session.

## Events

| Event | Payload | Producer | Consumer |
| --- | --- | --- | --- |
| `http:updated` | `Session`, including latest sample | `runner.go: Runner.emitLocked` → `main.go: main` callback → `App.Event.Emit` | `api.ts: onUpdate` → `useSessions.ts: merge` → `App.tsx`, `LatencyChart.tsx` |
| `http:removed` | Session ID string | `main.go: PingService.Remove` → `App.Event.Emit` | `api.ts: onRemove` → `useSessions.ts: useSessions` |

Updates occur on start, completed probe, and stop. Each carries a summary and at most one sample, not the whole history. Consumers ignore stale revisions. Selection subscribes before fetching detail and reconciles updates received during the snapshot. Removed IDs are ignored if older queued updates arrive.

## Start and results flow

```mermaid
flowchart TD
    A["frontend/src/App.tsx: start"] --> B["frontend/src/api.ts: api.start → Call.ByName"]
    B --> C["main.go: PingService.Start"]
    C --> D["internal/ping/runner.go: Runner.Start → validate"]
    R["frontend/src/App.tsx: restart"] --> S["frontend/src/api.ts: api.restart → Call.ByName"]
    S --> T["main.go: PingService.Restart"]
    T --> U["internal/ping/runner.go: Runner.Restart → startLocked"]
    U --> E
    D --> E["internal/ping/runner.go: Runner.run"]
    E --> F["internal/ping/runner.go: probe → http.Client.Do"]
    F --> G["internal/ping/runner.go: probe matches configured status ranges"]
    G --> V["internal/ping/runner.go: Runner.run updates counters and appends history"]
    V --> H["internal/ping/runner.go: Runner.emitLocked"]
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
    G --> H["internal/ping/runner.go: Runner.Get copies complete session history"]
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
