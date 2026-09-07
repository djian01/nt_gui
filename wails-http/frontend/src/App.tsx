import { useEffect, useState, type FormEvent } from "react";
import { motion } from "framer-motion";
import {
  Activity,
  ArrowDownUp,
  ArrowUpRight,
  Check,
  ChevronDown,
  ChevronRight,
  CircleHelp,
  Clock3,
  ExternalLink,
  Globe2,
  Layers3,
  Play,
  Plus,
  Radio,
  RotateCcw,
  Search,
  Server,
  ShieldCheck,
  SlidersHorizontal,
  Square,
  Timer,
  Trash2,
  X,
} from "lucide-react";
import { api, errorText, type Config, type Session } from "./api";
import { useSessions } from "./useSessions";
import LatencyChart, { clock, ms } from "./LatencyChart";

const chartID = new URLSearchParams(window.location.search).get("chart");
const loss = (s: Session) =>
  s.sent ? ((s.sent - s.succeeded) / s.sent) * 100 : 0;
const host = (s: Session) => new URL(s.config.url).host;
const stateLabel = (s: Session) =>
  !s.running
    ? "Stopped"
    : !s.last
      ? "Connecting"
      : s.last.success
        ? "Healthy"
        : "Failing";

function Badge({ session }: { session: Session }) {
  return (
    <span
      className={`status ${!session.running ? "stopped" : !session.last ? "pending" : session.last.success ? "healthy" : "failing"}`}
    >
      <i className="dot" />
      {stateLabel(session)}
    </span>
  );
}

export default function App() {
  const [selected, setSelected] = useState<string | null>(chartID);
  const { sessions, detail, connected, error, setError, merge } =
    useSessions(selected);
  const [config, setConfig] = useState<Config>({
    url: "",
    method: "GET",
    intervalMs: 1000,
    timeoutMs: 4000,
    acceptedStatuses: ["2xx", "3xx"],
    proxy: { enabled: false, url: "", username: "", password: "" },
  });
  const [scheme, setScheme] = useState<"https" | "http">("https");
  const [advanced, setAdvanced] = useState(false);
  const [customStatuses, setCustomStatuses] = useState("");
  const [busy, setBusy] = useState("");
  const [search, setSearch] = useState("");
  const [filter, setFilter] = useState("all");
  const [help, setHelp] = useState(false);
  const selectedSession = sessions.find((s) => s.id === selected) ?? null;
  const active = sessions.filter((s) => s.running).length;
  useEffect(() => {
    if (!chartID && (!selected || !sessions.some((s) => s.id === selected)))
      setSelected(sessions[0]?.id ?? null);
  }, [selected, sessions]);

  async function action(key: string, fn: () => Promise<unknown>) {
    setBusy(key);
    setError("");
    try {
      await fn();
    } catch (err) {
      setError(errorText(err));
    } finally {
      setBusy("");
    }
  }
  function start(e: FormEvent) {
    e.preventDefault();
    void action("start", async () => {
      const custom = customStatuses.split(/[\s,]+/).filter(Boolean);
      const s = await api.start({
        ...config,
        url: `${scheme}://${config.url.trim()}`,
        acceptedStatuses: [
          ...config.acceptedStatuses.filter((value) => value.endsWith("xx")),
          ...custom,
        ],
      });
      merge(s);
      setSelected(s.id);
    });
  }
  function setTarget(value: string) {
    const withScheme = value.match(/^\s*(https?):\/\/(.*)$/i);
    if (withScheme) {
      setScheme(withScheme[1].toLowerCase() as "https" | "http");
      setConfig({ ...config, url: withScheme[2] });
      return;
    }
    setConfig({ ...config, url: value.trimStart() });
  }
  const stop = (s: Session) =>
    action(s.id, async () => merge(await api.stop(s.id)));
  const restart = (s: Session) =>
    action(s.id, async () => {
      const next = await api.restart(s.id);
      merge(next);
      // A chart window keeps its original session identity for focus/removal.
      if (chartID) await api.chart(next.id);
      else setSelected(next.id);
    });
  const filtered = sessions.filter(
    (s) =>
      s.config.url.toLowerCase().includes(search.toLowerCase()) &&
      (filter === "all" || (filter === "running" ? s.running : !s.running)),
  );
  const session =
    detail?.session.id === selected ? detail.session : selectedSession;

  const metrics = (
    <div className="metrics">
      <Metric
        label="Latest response"
        value={session?.last ? ms(session.last.rtt) : "—"}
        unit={session?.last ? "ms" : ""}
        note={
          session?.last
            ? session.last.statusCode
              ? `HTTP ${session.last.statusCode} · ${session.last.success ? "Accepted" : "Failed"}`
              : "No HTTP response"
            : "Waiting for a probe"
        }
        icon={<Activity size={18} />}
      />
      <Metric
        label="Average latency"
        value={session?.succeeded ? ms(session.avgRtt) : "—"}
        unit={session?.succeeded ? "ms" : ""}
        note={
          session?.succeeded
            ? `Min ${ms(session.minRtt)} · Max ${ms(session.maxRtt)} ms`
            : "Successful responses only"
        }
        icon={<Timer size={18} />}
      />
      <Metric
        label="Failure rate"
        value={session?.sent ? loss(session).toFixed(1) : "—"}
        unit={session?.sent ? "%" : ""}
        note={
          session?.sent
            ? `${session.sent - session.succeeded} failed of ${session.sent} probes`
            : "Matches your expected statuses"
        }
        icon={<ArrowDownUp size={18} />}
        warning={!!session && loss(session) > 0}
      />
      <Metric
        label="Probes sent"
        value={session ? String(session.sent) : "0"}
        note={
          session
            ? `Every ${session.config.intervalMs / 1000}s · ${session.config.timeoutMs / 1000}s timeout`
            : "Select a test to view its metrics"
        }
        icon={<Radio size={18} />}
      />
    </div>
  );

  const analysis = (
    <div className={`analysis-grid ${chartID ? "detached" : ""}`}>
      <section className="panel chart-panel">
        <div className="panel-heading">
          <div>
            <h2>Response time</h2>
            <p>
              {session ? host(session) : "A clearer view of your connection"}
            </p>
          </div>
          <div className="heading-actions">
            {session?.running && (
              <span className="live-label">
                <i className="dot accent" />
                LIVE
              </span>
            )}
            {session && !chartID && (
              <button
                className="icon-button"
                title="Open chart in a separate window"
                aria-label="Open chart in a separate window"
                onClick={() =>
                  void action("chart", () => api.chart(session.id))
                }
              >
                <ExternalLink size={17} />
              </button>
            )}
          </div>
        </div>
        <LatencyChart
          samples={detail?.session.id === selected ? detail.samples : []}
        />
      </section>
      <section className="panel probes-panel">
        <div className="panel-heading">
          <div>
            <h2>Recent probes</h2>
            <p>The latest response, at a glance</p>
          </div>
          <span className="count-pill">
            {Math.min(6, detail?.samples.length ?? 0)}
          </span>
        </div>
        {detail?.samples.length ? (
          <div className="probe-list">
            {detail.samples
              .slice(-6)
              .reverse()
              .map((p) => (
                <div className="probe" key={p.sequence}>
                  <span className={`probe-icon ${p.success ? "ok" : "bad"}`}>
                    {p.success ? <Check size={14} /> : <X size={14} />}
                  </span>
                  <div>
                    <strong>
                      {p.statusCode ? `HTTP ${p.statusCode}` : "Request failed"}
                    </strong>
                    <span title={p.error || undefined}>
                      {p.error || `${clock(p.time)} · #${p.sequence}`}
                    </span>
                  </div>
                  <strong className="probe-rtt">
                    {ms(p.rtt)}
                    <small> ms</small>
                  </strong>
                </div>
              ))}
          </div>
        ) : (
          <div className="probes-empty">
            <Clock3 size={25} />
            <strong>Listening for results</strong>
            <p>Each completed probe will appear here.</p>
          </div>
        )}
      </section>
    </div>
  );

  return (
    <div className={`app ${chartID ? "chart-app" : ""}`}>
      {!chartID && (
        <aside className="sidebar">
          <div className="brand">
            <span className="brand-mark">
              <Activity size={25} strokeWidth={2.5} />
            </span>
            <div>
              <strong>NET TEST</strong>
              <span>Network diagnostics</span>
            </div>
          </div>
          <div className="workspace-label">WORKSPACE</div>
          <div className="sidebar-current">
            <Globe2 size={18} />
            <span>HTTP Ping</span>
            <ChevronRight size={15} />
          </div>
          <div className="sidebar-note">
            <Layers3 size={17} />
            <div>
              <strong>One protocol. A new view.</strong>
              <p>A focused desktop preview of the next Net Test interface.</p>
            </div>
          </div>
          <div className="sidebar-bottom">
            <button onClick={() => setHelp((v) => !v)}>
              <CircleHelp size={17} />
              About this preview
              <ArrowUpRight size={15} />
            </button>
            <div className="version">
              <i className="dot accent" />
              Wails desktop<span>0.1</span>
            </div>
          </div>
        </aside>
      )}
      <main>
        <header className="topbar">
          <div className="breadcrumb">
            Network tools
            <ChevronRight size={13} />
            <strong>{chartID ? "Latency detail" : "HTTP Ping"}</strong>
          </div>
          <span className={`connection ${connected ? "" : "disconnected"}`}>
            <i className="dot" />
            {connected ? "Desktop connected" : "Connecting to desktop"}
          </span>
        </header>
        <div className="page">
          <div className="page-heading">
            <div>
              <div className="eyebrow">
                {chartID ? "TEST INSIGHTS" : "OBSERVE YOUR ENDPOINTS"}
              </div>
              <h1>
                {chartID ? "Latency detail" : "HTTP Ping"}
                <span className="preview-tag">Preview</span>
              </h1>
              <p>
                {chartID
                  ? (session?.config.url ?? "Loading test…")
                  : "Measure responsiveness. Spot failures. Keep your network in view."}
              </p>
            </div>
            <div className="page-actions">
              {chartID && session ? (
                <>
                  <Badge session={session} />
                  <button
                    className="button secondary"
                    disabled={!!busy}
                    onClick={() =>
                      void (session.running ? stop(session) : restart(session))
                    }
                  >
                    {session.running ? (
                      <Square size={14} />
                    ) : (
                      <RotateCcw size={14} />
                    )}
                    {session.running ? "Stop test" : "Run again"}
                  </button>
                </>
              ) : (
                <span className="running-count">
                  <i className={`dot ${active ? "accent" : ""}`} />
                  {active} active {active === 1 ? "test" : "tests"}
                </span>
              )}
            </div>
          </div>
          {help && (
            <div className="help-box">
              <ShieldCheck size={20} />
              <div>
                <strong>About this HTTP preview</strong>
                <p>
                  GET, PUT, and PATCH probes use a fresh connection with system
                  TLS verification. Latency measures time to response headers.
                  Expected statuses are configurable, redirects are not
                  followed, and each test can use an HTTP or HTTPS proxy. Up to
                  8 tests can run at once. The latest 600 probes per test stay
                  in memory until you remove the test or quit. Database history
                  and other protocols are available in the existing Fyne app.
                </p>
              </div>
              <button
                className="icon-button"
                aria-label="Close preview information"
                onClick={() => setHelp(false)}
              >
                <X size={17} />
              </button>
            </div>
          )}
          {error && (
            <div className="error-banner" role="alert">
              <span>{error}</span>
              <button
                className="icon-button"
                onClick={() => setError("")}
                aria-label="Dismiss error"
              >
                <X size={16} />
              </button>
            </div>
          )}
          {!chartID && (
            <form className="test-form panel" onSubmit={start}>
              <div className="form-heading">
                <span className="section-icon">
                  <Plus size={16} />
                </span>
                <h2>New HTTP test</h2>
                <button
                  type="button"
                  className={`advanced-trigger ${advanced ? "open" : ""}`}
                  aria-expanded={advanced}
                  onClick={() => setAdvanced((value) => !value)}
                >
                  <SlidersHorizontal size={13} />
                  Advanced
                  <ChevronDown size={13} />
                </button>
              </div>
              <div className="form-fields">
                <label className="url-field">
                  Target
                  <div className="url-composer">
                    <select
                      className="scheme-select"
                      aria-label="URL protocol"
                      value={scheme}
                      onChange={(event) =>
                        setScheme(event.target.value as "https" | "http")
                      }
                    >
                      <option value="https">HTTPS</option>
                      <option value="http">HTTP</option>
                    </select>
                    <span aria-hidden="true">://</span>
                    <input
                      type="text"
                      required
                      maxLength={4088}
                      placeholder="example.com/path"
                      value={config.url}
                      onChange={(event) => setTarget(event.target.value)}
                      spellCheck={false}
                      autoCapitalize="none"
                    />
                  </div>
                </label>
                <label>
                  Method
                  <select
                    value={config.method}
                    onChange={(e) =>
                      setConfig({ ...config, method: e.target.value })
                    }
                  >
                    <option>GET</option>
                    <option>PUT</option>
                    <option>PATCH</option>
                  </select>
                </label>
                <label>
                  Interval
                  <select
                    value={config.intervalMs}
                    onChange={(e) =>
                      setConfig({
                        ...config,
                        intervalMs: Number(e.target.value),
                      })
                    }
                  >
                    {[1, 2, 5, 10, 30, 60].map((n) => (
                      <option value={n * 1000} key={n}>
                        {n} second{n === 1 ? "" : "s"}
                      </option>
                    ))}
                  </select>
                </label>
                <label>
                  Timeout
                  <select
                    value={config.timeoutMs}
                    onChange={(e) =>
                      setConfig({
                        ...config,
                        timeoutMs: Number(e.target.value),
                      })
                    }
                  >
                    {[1, 2, 4, 10, 30].map((n) => (
                      <option value={n * 1000} key={n}>
                        {n} second{n === 1 ? "" : "s"}
                      </option>
                    ))}
                  </select>
                </label>
                <button
                  className="button primary start-button"
                  type="submit"
                  disabled={!!busy || !connected || active >= 8}
                >
                  <Play size={15} fill="currentColor" />
                  {busy === "start" ? "Starting…" : "Start test"}
                </button>
              </div>
              {advanced && (
                <motion.div
                  className="advanced-panel"
                  initial={{ opacity: 0, height: 0 }}
                  animate={{ opacity: 1, height: "auto" }}
                  transition={{ duration: 0.2 }}
                >
                  <div className="advanced-section status-options">
                    <div className="advanced-copy">
                      <span className="advanced-icon">
                        <ShieldCheck size={16} />
                      </span>
                      <div>
                        <strong>Expected HTTP statuses</strong>
                        <p>Only selected responses count as successful.</p>
                      </div>
                    </div>
                    <div className="status-controls">
                      <div className="status-groups">
                        {["2xx", "3xx", "4xx", "5xx"].map((group) => (
                          <label className="check-pill" key={group}>
                            <input
                              type="checkbox"
                              checked={config.acceptedStatuses.includes(group)}
                              onChange={(event) =>
                                setConfig({
                                  ...config,
                                  acceptedStatuses: event.target.checked
                                    ? [...config.acceptedStatuses, group]
                                    : config.acceptedStatuses.filter(
                                        (value) => value !== group,
                                      ),
                                })
                              }
                            />
                            <span>{group}</span>
                          </label>
                        ))}
                      </div>
                      <label className="custom-status">
                        Custom codes
                        <input
                          type="text"
                          inputMode="numeric"
                          placeholder="e.g. 404, 429"
                          value={customStatuses}
                          onChange={(event) =>
                            setCustomStatuses(event.target.value)
                          }
                        />
                      </label>
                    </div>
                  </div>
                  <div className="advanced-divider" />
                  <div className="advanced-section proxy-options">
                    <div className="advanced-copy">
                      <span className="advanced-icon">
                        <Server size={16} />
                      </span>
                      <div>
                        <strong>HTTP proxy</strong>
                        <p>Route this test through an HTTP or HTTPS proxy.</p>
                      </div>
                    </div>
                    <label className="switch-control">
                      <input
                        type="checkbox"
                        role="switch"
                        checked={config.proxy.enabled}
                        onChange={(event) =>
                          setConfig({
                            ...config,
                            proxy: {
                              ...config.proxy,
                              enabled: event.target.checked,
                            },
                          })
                        }
                      />
                      <span className="switch-track">
                        <i />
                      </span>
                      {config.proxy.enabled ? "Enabled" : "Off"}
                    </label>
                  </div>
                  {config.proxy.enabled && (
                    <div className="proxy-fields">
                      <label className="proxy-url">
                        Proxy URL
                        <input
                          type="url"
                          required
                          placeholder="http://proxy.example:8080"
                          value={config.proxy.url}
                          onChange={(event) =>
                            setConfig({
                              ...config,
                              proxy: {
                                ...config.proxy,
                                url: event.target.value,
                              },
                            })
                          }
                        />
                      </label>
                      <label>
                        Username
                        <input
                          autoComplete="off"
                          placeholder="Optional"
                          value={config.proxy.username}
                          onChange={(event) =>
                            setConfig({
                              ...config,
                              proxy: {
                                ...config.proxy,
                                username: event.target.value,
                              },
                            })
                          }
                        />
                      </label>
                      <label>
                        Password
                        <input
                          type="password"
                          autoComplete="new-password"
                          placeholder="Optional"
                          value={config.proxy.password}
                          onChange={(event) =>
                            setConfig({
                              ...config,
                              proxy: {
                                ...config.proxy,
                                password: event.target.value,
                              },
                            })
                          }
                        />
                      </label>
                    </div>
                  )}
                  <div className="advanced-note">
                    <ShieldCheck size={13} /> Proxy passwords are used by the Go
                    runner and are not returned to the interface.
                  </div>
                </motion.div>
              )}
            </form>
          )}
          {metrics}
          {!chartID && (
            <section className="panel sessions-panel">
              <div className="panel-heading">
                <div className="title-with-count">
                  <h2>Your tests</h2>
                  <span className="count-pill">{sessions.length}</span>
                </div>
                <div className="table-tools">
                  <div className="segmented">
                    <button
                      onClick={() => setFilter("all")}
                      aria-pressed={filter === "all"}
                    >
                      All
                    </button>
                    <button
                      onClick={() => setFilter("running")}
                      aria-pressed={filter === "running"}
                    >
                      Running
                    </button>
                    <button
                      onClick={() => setFilter("stopped")}
                      aria-pressed={filter === "stopped"}
                    >
                      Stopped
                    </button>
                  </div>
                  <label className="search">
                    <Search size={14} />
                    <input
                      aria-label="Filter tests by URL"
                      placeholder="Filter endpoints…"
                      value={search}
                      onChange={(e) => setSearch(e.target.value)}
                    />
                  </label>
                </div>
              </div>
              <div className="table-scroll">
                <table>
                  <thead>
                    <tr>
                      <th>ENDPOINT</th>
                      <th>STATUS</th>
                      <th>LATEST</th>
                      <th>AVERAGE</th>
                      <th>FAILURE</th>
                      <th>PROBES</th>
                      <th className="actions-cell">ACTIONS</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filtered.map((s) => (
                      <tr
                        key={s.id}
                        className={s.id === selected ? "selected" : ""}
                        tabIndex={0}
                        aria-label={`Select ${s.config.url} for the live graph`}
                        aria-selected={s.id === selected}
                        onClick={() => setSelected(s.id)}
                        onKeyDown={(event) => {
                          if (
                            event.target === event.currentTarget &&
                            (event.key === "Enter" || event.key === " ")
                          ) {
                            event.preventDefault();
                            setSelected(s.id);
                          }
                        }}
                      >
                        <td>
                          <div className="endpoint-button">
                            <span className="endpoint-icon">
                              <Globe2 size={17} />
                            </span>
                            <span>
                              <strong title={s.config.url}>
                                {s.config.url}
                              </strong>
                              <small>
                                {s.config.method} <span>·</span> Started{" "}
                                {clock(s.startedAt)}
                              </small>
                            </span>
                          </div>
                        </td>
                        <td>
                          <Badge session={s} />
                        </td>
                        <td className="numeric">
                          {s.last ? `${ms(s.last.rtt)} ms` : "—"}
                        </td>
                        <td className="numeric">
                          {s.succeeded ? `${ms(s.avgRtt)} ms` : "—"}
                        </td>
                        <td
                          className={`numeric ${loss(s) ? "text-failure" : ""}`}
                        >
                          {s.sent ? `${loss(s).toFixed(1)}%` : "—"}
                        </td>
                        <td className="numeric">{s.sent}</td>
                        <td>
                          <div
                            className="row-actions"
                            onClick={(event) => event.stopPropagation()}
                          >
                            <button
                              className="icon-button"
                              title="Open chart"
                              aria-label={`Open chart for ${s.config.url}`}
                              disabled={!!busy}
                              onClick={() =>
                                void action("chart", () => api.chart(s.id))
                              }
                            >
                              <ExternalLink size={15} />
                            </button>
                            <button
                              className={`icon-button ${s.running ? "stop-action" : ""}`}
                              title={s.running ? "Stop test" : "Run again"}
                              aria-label={`${s.running ? "Stop" : "Run again"} ${s.config.url}`}
                              disabled={!!busy || (!s.running && active >= 8)}
                              onClick={() =>
                                void (s.running ? stop(s) : restart(s))
                              }
                            >
                              {s.running ? (
                                <Square size={13} fill="currentColor" />
                              ) : (
                                <RotateCcw size={15} />
                              )}
                            </button>
                            <button
                              className="icon-button"
                              title={
                                s.running
                                  ? "Stop before removing"
                                  : "Remove test"
                              }
                              aria-label={`Remove ${s.config.url}`}
                              disabled={s.running || !!busy}
                              onClick={() =>
                                void action(s.id, () => api.remove(s.id))
                              }
                            >
                              <Trash2 size={15} />
                            </button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              {!filtered.length && (
                <div className="table-empty">
                  <span className="endpoint-icon">
                    <Globe2 size={21} />
                  </span>
                  <div>
                    <strong>
                      {sessions.length
                        ? "No matching endpoints"
                        : "No tests yet. Let’s check your first endpoint."}
                    </strong>
                    <p>
                      {sessions.length
                        ? "Try another filter or search term."
                        : "Enter a URL above and start a test. Your live results will appear here."}
                    </p>
                  </div>
                </div>
              )}
              <div className="table-footer">
                <span>
                  {selectedSession ? (
                    <>
                      Viewing <strong>{host(selectedSession)}</strong>
                    </>
                  ) : (
                    "Select an endpoint to inspect its response times"
                  )}
                </span>
                <span>{sessions.length} / 24 sessions</span>
              </div>
            </section>
          )}
          {analysis}
          <footer className="page-footer">
            <span>
              <ShieldCheck size={14} />
              {chartID
                ? "Live session · shared with your main window"
                : "Runs locally on your desktop"}
            </span>
            <span>
              Session results stay in memory · Last 600 probes retained
            </span>
          </footer>
        </div>
      </main>
    </div>
  );
}

function Metric({
  label,
  value,
  unit = "",
  note,
  icon,
  warning = false,
}: {
  label: string;
  value: string;
  unit?: string;
  note: string;
  icon: React.ReactNode;
  warning?: boolean;
}) {
  return (
    <motion.div
      className={`metric ${warning ? "warning" : ""}`}
      initial={{ opacity: 0, y: 5 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.22 }}
    >
      <div className="metric-label">
        {label}
        {icon}
      </div>
      <div className="metric-value">
        {value}
        <span>{unit}</span>
      </div>
      <div className="metric-note">{note}</div>
    </motion.div>
  );
}
