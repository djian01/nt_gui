# HTTP function parity: Fyne → Wails

Reviewed 2026-09-08, continued 2026-09-09, against the root Fyne working tree (nt v1.4.0) and `desktop/`. Two requested GPT-5.6 Sol agents with high reasoning audited the legacy and current implementations. They returned findings/partial changes before reaching their usage limits; the primary agent completed integration and verification. This compares HTTP behavior and the shared tools usable by HTTP. TCP/DNS/ICMP implementation is outside this review.

## Result and resolved gaps

- Restored POST alongside GET/PUT; retained Wails PATCH as an additional method.
- Removed the Wails-only 60-second interval / 30-second timeout caps; input remains bounded by the supported Go duration to avoid overflow.
- Added optional redirect following to reproduce legacy final-response probing. Existing Wails sessions keep first-response probing by default.
- Preserve the response reason phrase in new results, CSV exports, and legacy CSV imports.
- Added HTTP CSV import and analysis, including legacy files whose recording began mid-test. Imports are stopped, independently saved sessions with summary, zoom, replay, CSV export, and chart export available.
- Extended CSV exports with expected statuses, proxy URL/username/password-required flag, redirect policy, and reason phrase. Proxy passwords remain excluded. The first eleven existing columns retain their order.
- Added visible-chart PNG export.
- Added selection, page select-all, confirmed bulk deletion, and selection recovery after another window removes a session.
- Added stopped end time/duration and import context to the analysis view.

## Function mapping

Wails paths below are relative to `desktop/`; Fyne paths are relative to the repository root. Multiple legacy functions are grouped only when they implement the same behavior or toolkit rendering contract.

| Legacy file and function(s) | HTTP responsibility | Current file and function(s) | Outcome |
| --- | --- | --- | --- |
| `HTTP_Ping_UI.go: HTTPPingContainer` | New HTTP test action, table and scrolling | `frontend/src/App.tsx: App / start` | Covered |
| `Function_NewTest.go: NewTest` HTTP branch | Scheme, target, methods, interval, timeout, status groups, custom status, proxy, recording | `App.tsx: start / setTarget`, `runner.go: validate / validateStatuses / Start` | Covered; POST and longer timings restored; recording automatic |
| `HTTP_Ping_Func_Struct.go: httpGUIRow.Initial / GenerateHeaderRow` | Row controls and columns | `App.tsx: App` table, `Badge`, metrics | Covered by React layout |
| `HTTP_Ping_Func_Struct.go: httpGUIRow.UpdateRow` | Sequence, status, URL/method, status code, RTT, start, failure, average, recording | `runner.go: run`, `useSessions.ts: merge / append`, `App.tsx: metrics / table / recent probes` | Covered; saved status replaces recording toggle |
| `HTTP_Ping_Func_Struct.go: httpObject.GetType / GetSummary / GetChartData / GetUUID` | Protocol, summary, timeline and identity access | `runner.go: Session / Get`, `store.go: Timeline`, `api.ts: Session / Detail / Timeline` | Covered through serialized models and bounded queries |
| `HTTP_Ping_Func_Struct.go: httpObject.Initial` | Initialize test state and row | `runner.go: startLocked`, `App.tsx: start` | Covered |
| `HTTP_Ping_Func_Struct.go: httpObject.UpdateRecording` | Enable recording after start | `runner.go: run → Store.Save` | Outcome covered by always-on persistence; toggle intentionally absent |
| `HTTP_Ping_Func_Struct.go: httpObject.UpdateChartData` | Append chart point | `store.go: Save / saveSample / Timeline`, `LatencyChart.tsx: range effect` | Covered without retaining full history in frontend memory |
| `HTTP_Ping_Func_Struct.go: httpObject.DisplayChartDataTerminal` | Debug chart output | No product equivalent | Debug-only helper; not a user feature |
| `HTTP_Ping_Func_Struct.go: httpObject.Stop` | Stop worker, end summary, enable replay | `runner.go: Stop`, `App.tsx: stop / restart` | Covered; cancellation aborts in-flight request |
| `HTTP_Ping_Func_Struct.go: HttpAddPingRow` | Worker lifecycle, callbacks, stop/replay/close/chart actions | `runner.go: Start / startLocked / run / Stop / Restart / Remove`, `App.tsx`, `main.go: OpenChart` | Covered; history deletion differs from hiding a row |
| `HTTP_Ping_Func_Struct.go: ConstructURL` | Construct scheme/host/port/path | `App.tsx: setTarget / start`, `runner.go: validate` | Covered; preserves query strings |
| `HTTP_Ping_Func_Struct.go: BuildProxyURL` | Build and validate proxy URL and credentials | `runner.go: validate`, `App.tsx: proxy fields` | Covered with explicit proxy URL input |
| `Function_Shared.go: ParseURL2HttpVars / ParseURL / ParseTargetURL / parseURL` | Parse HTTP target, URL and default port | `runner.go: validate`, Go `net/url`, `App.tsx: setTarget` | Covered; DNS resolution deferred to actual probes |
| `Function_Shared.go: ValidateAndResolve / targetHostValidator / IsValidIP` | Host/address syntax and optional resolution | `runner.go: validate / probe`, standard HTTP transport | Relevant HTTP outcome covered; no eager DNS requirement |
| `Function_Shared.go: Iv2NtCmd` | Reproduce test as nt command | No current equivalent | CLI command copying was removed from the desktop app |
| `Function_Shared.go: NtCmd2Iv` | Rebuild config for replay | `runner.go: Restart` uses saved `Config` | Covered through typed config, no lossy command reparse |
| `Function_Shared.go: existingTestCheck / UnregisterTest / GenerateShortUUID` | Lifecycle identity bookkeeping | `runner.go: sessions / Start / Stop / Remove / Close`, `crypto/rand.Text` | Covered |
| `Function_Shared.go: CloneChartPoints` | Stable chart snapshot | `store.go: Timeline` transaction, `LatencyChart.tsx: pause/range state` | Covered without full dataset cloning |
| `Function_Shared.go: getConfigFilePath` | Resolve application storage location | `store.go: DefaultStorePath / OpenStore` | Covered with separate Wails database |
| `Function_Shared.go: GUIVerticalSeparator / TruncateString / formCell / placeHolderBlock / getPrimaryScreenSize` | Fyne layout, truncation and window sizing | `App.tsx / style.css`, `main.go: window options` | Toolkit-specific presentation replaced |
| `Function_NewChartWindow.go: NewChartWindow` | Detached chart, stop, replay, recording, CSV export, close | `main.go: OpenChart / ExportCSV`, `App.tsx`, `LatencyChart.tsx` | Covered; one native chart window per session |
| `Function_NewChartWindow.go: NewChartUpdate` | Periodic chart/summary refresh and pause | `useSessions.ts: merge`, `LatencyChart.tsx: effects / pause / resetZoom` | Covered; pause does not stop recording |
| `Structs_Shared.go: Chart.Initial / ChartUpdate` | Initialize/render chart | `LatencyChart.tsx: LatencyChart / smoothPath` | Covered with interactive SVG |
| `Structs_Shared.go: SummaryData.Initial / UpdateRunning / GetUUID` | Live summary and identity | `runner.go: Session / startLocked / run` | Covered |
| `Structs_Shared.go: DbTestEntry2SummaryData` | Restore historical summary | `store.go: Get`, `runner.go: Get` | Covered without reconstructing entire raw history |
| `Structs_Shared.go: SummaryUI.Initial / CreateCard / UpdateStaticUI / UpdateUI_Running / UpdateUI_Ended` | Summary cards, command, start/end, counts, loss and RTT | `App.tsx: Metric / metrics / chart-test-meta / duration` | Summary covered; CLI command copying intentionally absent |
| `Structs_Shared.go: Slider.Initial / CreateCard / update / BuildSliderChartData / UpdateChartImage / ResetChartImage` | Select chart period, apply, reset | `LatencyChart.tsx: range inputs / range effect / resetZoom`, `store.go: Timeline` | Covered; changing handles applies automatically |
| `pkg/ntwidget/RangeSlider.go: NewRangeSlider / UpdateValues / CreateRenderer / Layout / MouseDown / MouseUp / Dragged / DragEnd / detectHandle / valueToPosition / positionToValue / updateGraphics` and renderer methods | Two-handle input and rendering | `LatencyChart.tsx: native HTML range inputs`, `style.css: dual-range` | Toolkit machinery replaced; zoom capability covered |
| `pkg/chart/chart.go: ConvertFromPacketToChartPoint / ConvertFromDbTestEntryToChartPoint / ConvertFromDbToCheckpoint / CreateChart / createPlaceholderImage / drawCenteredLabel` | Chart conversion, drawing, empty state | `runner.go: Sample`, `store.go: Timeline`, `LatencyChart.tsx` | Covered; representative points for dense history |
| `Function_CSVExport.go: SaveToCSV / GetDefaultExportFolder` | Export raw HTTP results to selected file | `store.go: ExportCSV`, `main.go: ExportCSV / saveExport` | Covered; streaming, fixed watermark, native dialog |
| `History_UI.go: HistoryContainer` | History list, filter, select-all, bulk delete | `App.tsx: History view / checked / bulkDeleting`, `useSessions.ts`, `store.go: List` | Covered for HTTP; select-all applies to current page |
| `History_Func_Struct.go: historyGUIRow.Initial / GenerateHeaderRow / UpdateRow` | History rows and actions | `App.tsx: table / Badge / restart` | Covered except intentionally removed CLI command copying |
| `History_Func_Struct.go: historyAddRow / historyRefresh / createTestObj / GetHistoryEntries` | Load and display saved tests/charts | `useSessions.ts: list / selection effects`, `runner.go: List / Get`, `main.go: OpenChart` | Covered with pagination |
| `History_Func_Struct.go: AddSelectedEntry / DelSelectedEntry / EntryExist / selectOperation` | Multi-selection bookkeeping | `App.tsx: checked / checkbox handlers / bulk deletion` | Covered |
| `Result_Analysis_UI.go: ResultAnalysisContainer / OpenResultCSVFile` | Choose existing results, analyze, replay | `App.tsx: import action / saved analysis / runAgain`, `main.go: ImportCSV`, `runner.go: ImportCSV` | Covered for HTTP CSV |
| `Result_Analysis_Func.go: appendPacket` HTTP branch | Parse CSV samples and summary | `import.go: Store.ImportCSV / parseLegacyHTTPRow / parseWailsRow / updateImportedSummary` | Covered with validation and rollback |
| `Result_Analysis_Func.go: parseCustomDuration / RaNtCmdGenerator` | Parse legacy timing and construct replay | `import.go: durationMilliseconds / legacyTime`, `runner.go: Restart` | In-app replay covered; CLI command generation intentionally absent; missing source configuration explicitly disclosed |
| `pkg/ntdb/ntdb.go: DBOpen / createDatabase / createHistoryTable / CreateTestResultsTable` | Initialize persistent storage | `store.go: OpenStore / initialize` | Covered with one shared schema; no Fyne DB migration |
| `pkg/ntdb/ntdb.go: ConvertPkt2DbEntry / InsertEntry / UpdateFieldValue` | Persist probes, metadata, recording state | `store.go: Save / saveSample`, `runner.go: run / Stop` | Covered atomically before UI events |
| `pkg/ntdb/ntdb.go: ReadTableEntries / ReadTestTableEntries / SortHistoryEntries / ConvertDbEntriesToHistoryEntries / ConvertDbTestEntriesToRecordHTTPEntries` | History/raw reads and conversions | `store.go: List / Get / Recent / Timeline / ExportCSV` | Covered with indexed/paged queries |
| `pkg/ntdb/ntdb.go: DeleteEntry / DeleteTable` | Delete history and recorded results | `store.go: Remove`, foreign-key cascades | Covered for stopped tests |
| `pkg/ntdb/ntdb.go: isBusyOrLocked / queryWithRetry / execWithRetry` | Database contention handling | `store.go: initialize` busy timeout / single connection, `runner.go: ImportCSV` | Different mechanism; errors surfaced and live imports disallowed |
| `pkg/ntdb/ntdb.go: ShowHistoryTableConsole` | Debug history output | No product equivalent | Debug-only helper |

## Deliberate differences and limits

1. Wails verifies target/proxy TLS certificates. The legacy nt HTTP adapter disables verification. This audit does not copy that insecure default.
2. Wails defaults to the first HTTP response. Enable Follow redirects for legacy final-response behavior. POST/PUT/PATCH still have empty request bodies; neither legacy form nor this form supports arbitrary body/header editing.
3. All Wails probes are saved. No recording-off mode or removal of a live row while retaining separate hidden history. Eight active sessions remain the existing Wails resource limit.
4. Wails reports DNS failures as test results instead of rejecting a test during form validation. Proxy configuration requires a complete HTTP/HTTPS URL rather than separate host/port fields and inferred defaults.
5. The chart shows exact range statistics and bounded representative points. Its pause/zoom behavior is automatic and does not use Fyne's separate apply button. Historical raw probes remain available through zoom/CSV.
6. CSV import supports one HTTP test per file, chronological timestamps, consecutive source sequences, at most 256 MiB / one million rows. Partial legacy recordings are renumbered from one, and summary statistics cover the imported subset. Original starting sequence is disclosed in the import note.
7. Legacy CSV omits interval, timeout, expected-status and proxy settings. Replay defaults to 1s/4s, follows redirects, has no proxy, and infers expected statuses from successful samples. Earlier Wails exports preserve timings but lack proxy/status/redirect metadata. The interface displays these limitations. A proxy password is never recovered from CSV.
8. Legacy timezone abbreviations are interpreted using the local timezone where recognized, then conventional fixed offsets for supported abbreviations. Ambiguous abbreviations remain inherently lossy; unknown zones are rejected. Current CSV exports use UTC timestamps.
9. Import requires stopping active tests before its atomic transaction. Bulk deletion is limited to selected stopped tests on the current page and commits each deletion independently; successful deletions are removed from retry selection.
10. PNG export captures the visible chart, endpoint, method, range and statistics. It is a chart image, not a screenshot of the entire desktop. The desktop app does not generate or copy nt CLI commands.
11. Windows/Linux interactive verification, arbitrary clock-backward scenarios, million-row import benchmarks, extended soak tests and import/export progress cancellation remain future work. None are claimed as completed by local build success.

## Verification

- Legacy baseline: `go test -tags ci ./...` succeeds (compilation; no legacy test files).
- Desktop: race-enabled Go suite passes for methods/statuses, proxy authentication and password redaction, TLS rejection, timeout/cancellation, replay, persistence/reopen, 4,200-probe overview/range/export, history pagination/deletion and transactional failure.
- Added regression cases: POST and redirect policy survive restart; long timings accepted with overflow rejected; current CSV configuration round-trip; legacy HTTP timezone/reason phrase; partial legacy recordings; invalid-row rollback; changed/credential-bearing proxy metadata rejection; import refused while live tests run; PNG validation; failed export preserves destination.
- Frontend: TypeScript/Vite production build passes.
- Native macOS production application builds with matching deployment targets. In an isolated temporary database, the native CSV picker imported a legacy partial recording beginning at sequence 10: 20 probes, 16 successful, 20% failure, 20ms average, POST, AEST timestamps, reason phrases, and an explicit missing-settings note were verified. The native Save PNG dialog wrote a valid 1800×690 image; visual inspection verified its plot, failure markers, axes, endpoint, method, statistics and range.
- CSV save exported twenty ordered rows with replay metadata; cancelling the save dialog created no file. Re-importing that current-format CSV produced a second stopped session with the same settings/statistics.
- Native history select-all selected both rows and displayed the two-test confirmation. Cancellation preserved both tests; selections were cleared afterward. Actual destructive database deletion is covered by automated lifecycle/store tests, not by deleting the user's results. The normal app was restored with the original two saved tests intact; isolated verification files were removed.

API contracts and file/function flow charts: [desktop API](../api_reference/desktop.md). Storage/rendering architecture: [desktop architecture](desktop.md).


## Follow-up: shared recording and table format

HTTP now exposes the same new-test Recording checkbox as DNS (off by default). The shared `recording.go` implementation keeps live charts available while off, saves subsequent probes when Record is enabled, and preserves existing HTTP recordings and missing-type API compatibility. New HTTP/DNS/History rows use the same Endpoint / Status / Latest / Average / Failure / Probes / Actions layout, with actions at the right. This supersedes earlier always-save-only HTTP behavior.
