import { useEffect, useRef, useState, type FormEvent, type ReactNode } from "react";
import { motion } from "framer-motion";
import {
  Activity,
  History,
  ArrowDownUp,
  ArrowUpRight,
  Check,
  ChevronDown,
  ChevronRight,
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
  Moon,
  Sun,
  Square,
  Timer,
  Trash2,
  Upload,
  X,
} from "lucide-react";
import { Browser } from "@wailsio/runtime";
import { api, errorText, onRemove, type Config, type Session } from "./api";
import { useSessions } from "./useSessions";
import { DNSForm } from "./DNSTest";
import { TCPForm } from "./TCPTest";
import { ICMPForm } from "./ICMPTest";
import LatencyChart, { clock, ms } from "./LatencyChart";

const chartID = new URLSearchParams(window.location.search).get("chart");
type Theme = "dark" | "light";
const themePreferenceKey = "net-test-theme";
const initialTheme = (): Theme => {
  const saved = localStorage.getItem(themePreferenceKey);
  if (saved === "dark" || saved === "light") return saved;
  return window.matchMedia("(prefers-color-scheme: light)").matches ? "light" : "dark";
};
const loss = (s: Session) =>
  s.sent ? ((s.sent - s.succeeded) / s.sent) * 100 : 0;
const isTCP = (s: Session) => s.config.type === "tcp";
const isICMP = (s: Session) => s.config.type === "icmp";
const isNetwork = (s: Session) => isDNS(s) || isTCP(s) || isICMP(s);
const isDNS = (s: Session) => s.config.type === "dns";
const isRecording = (s: Session) => !s.config.type || !!s.config.recording;
const dnsResult = (s: Session) => s.last?.dnsResponse || s.last?.error || (s.last?.success ? "No IPv4 answer" : s.last ? "No response" : "Waiting for a probe");
const shortDNSResult = (s: Session) => { const value = dnsResult(s); return value.length > 40 ? `${value.slice(0, 39)}…` : value; };
const host = (s: Session) => (isTCP(s) || isICMP(s)) ? s.config.target : isDNS(s) ? s.config.resolver : new URL(s.config.url).host;
const duration = (s: Session) => {
  const elapsed = Math.max(0, Date.parse(s.endedAt ?? new Date().toISOString()) - Date.parse(s.startedAt));
  const seconds = Math.floor(elapsed / 1000);
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  return hours ? `${hours}h ${minutes}m` : minutes ? `${minutes}m ${seconds % 60}s` : `${seconds}s`;
};
const stateLabel = (s: Session) =>
  !s.running
    ? s.endReason === "interrupted" ? "Interrupted" : s.endReason === "storage_error" ? "Saving failed" : s.endReason === "imported" ? "Imported" : "Stopped"
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
  const [theme, setTheme] = useState<Theme>(initialTheme);
  const [selected, setSelected] = useState<string | null>(chartID);
  const [search, setSearch] = useState("");
  const [historyType, setHistoryType] = useState("all");
  const detailsRef = useRef<HTMLDivElement>(null);
  const [rowSelection, setRowSelection] = useState(0);

  function selectTest(id: string) {
    setSelected(id);
    setRowSelection(value => value + 1);
  }

  useEffect(() => {
    if (!rowSelection) return;
    detailsRef.current?.scrollIntoView({
      behavior: window.matchMedia("(prefers-reduced-motion: reduce)").matches ? "instant" : "smooth",
      block: "start",
    });
  }, [rowSelection]);
  const [cursors, setCursors] = useState<number[]>([0]);
  const [view, setView] = useState<"http" | "dns" | "tcp" | "icmp" | "history">("http");
  const filter = view === "history" ? (historyType === "all" ? "stopped" : `stopped-${historyType}`) : `current-${view}`;
  const { sessions, detail, connected, error, setError, merge, refresh, next, active, loading } =
    useSessions(selected, search, filter, cursors.at(-1) ?? 0);
  const filtered = view === "history" ? sessions.filter((s) => !s.running) : sessions;
  const [config, setConfig] = useState<Config>({
    type: "http",
    recording: false,
    url: "",
    method: "GET",
    intervalMs: 1000,
    timeoutMs: 4000,
    acceptedStatuses: ["2xx", "3xx"],
    followRedirects: false,
    proxy: { enabled: false, url: "", username: "", password: "" },
  });
  const [scheme, setScheme] = useState<"https" | "http">("https");
  const [advanced, setAdvanced] = useState(false);
  const [customStatuses, setCustomStatuses] = useState("");
  const [busy, setBusy] = useState("");
  const [aboutOpen, setAboutOpen] = useState(false);
  const selectedSession = filtered.find((s) => s.id === selected) ?? null;
  const [notice, setNotice] = useState("");
  const [deleting, setDeleting] = useState<Session | null>(null);
  const [bulkDeleting, setBulkDeleting] = useState<string[]>([]);
  const [checked, setChecked] = useState<Set<string>>(new Set());
  const [replaying, setReplaying] = useState<Session | null>(null);
  const [password, setPassword] = useState("");
  useEffect(() => {
    document.documentElement.dataset.theme = theme;
    document.documentElement.style.colorScheme = theme;
    localStorage.setItem(themePreferenceKey, theme);
  }, [theme]);
  useEffect(() => {
    const syncTheme = (event: StorageEvent) => {
      if (event.key === themePreferenceKey && (event.newValue === "dark" || event.newValue === "light"))
        setTheme(event.newValue);
    };
    window.addEventListener("storage", syncTheme);
    return () => window.removeEventListener("storage", syncTheme);
  }, []);
  useEffect(() => setCursors([0]), [search, filter, view]);
  useEffect(() => {
    setChecked((current) =>
      new Set([...current].filter((id) => sessions.some((s) => s.id === id && !s.running))),
    );
  }, [sessions]);
  useEffect(() => {
    if (chartID) return;
    const current = detail?.session.id === selected
      ? detail.session
      : sessions.find((s) => s.id === selected);
    if (view === "history") {
      if (current?.running) setSelected(null);
      return;
    }
    if (!selected)
      setSelected(sessions[0]?.id ?? null);
  }, [selected, sessions, detail, view]);
  useEffect(() => onRemove((id) => setSelected((current) => current === id ? null : current)), []);

  async function action(key: string, fn: () => Promise<unknown>) {
    setBusy(key);
    setError("");
    setNotice("");
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
      setCursors([0]); refresh();
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
  const runAgain = (s: Session, secret = "") =>
    action(s.id, async () => {
      const next = await api.restart(s.id, secret);
      merge(next);
      // A chart window keeps its original session identity for focus/removal.
      if (chartID) await api.chart(next.id);
      else { setView(isICMP(next) ? "icmp" : isTCP(next) ? "tcp" : isDNS(next) ? "dns" : "http"); setSelected(next.id); }
      setReplaying(null); setPassword(""); setCursors([0]); refresh();
    });
  const restart = (s: Session) => {
    if (s.passwordRequired) { setPassword(""); setReplaying(s); return Promise.resolve(); }
    return runAgain(s);
  };
  const selectedDetail =
    detail?.session.id === selected ? detail.session : selectedSession;
  const session = chartID || view !== "history" || selectedDetail?.running === false
    ? selectedDetail
    : null;
  const samples = session && detail?.session.id === session.id ? detail.samples : [];
  const protocol = session
    ? isICMP(session) ? "ICMP" : isTCP(session) ? "TCP" : isDNS(session) ? `DNS · ${session.config.protocol?.toUpperCase()}` : new URL(session.config.url).protocol.slice(0, -1).toUpperCase()
    : "";

  const metrics = (
    <div className="metrics" ref={detailsRef}>
      <Metric
        label="Latest response"
        value={session?.last && (!isNetwork(session) || session.last.success) ? ms(session.last.rtt) : "—"}
        unit={session?.last && (!isNetwork(session) || session.last.success) ? "ms" : ""}
        note={
          session?.last
            ? isICMP(session) ? (session.last.success ? "Echo reply" : session.last.error || "No reply") : isTCP(session) ? (session.last.success ? "Connected" : session.last.error || "Connection failed") : isDNS(session) ? (session.last.dnsResponse || session.last.error || (session.last.success ? "Success" : "Fail")) : session.last.statusCode
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
    <div className={`analysis-grid ${chartID ? "detached" : ""} ${view === "history" ? "history-detail" : ""}`}>
      <section className="panel chart-panel">
        <div className="panel-heading">
          <div className="chart-heading-copy">
            <div className="chart-title-line">
              <h2>Response time</h2>
              {session && (
                <>
                  <span className="test-chip protocol-chip">{protocol}</span>
                  <span className="test-chip method-chip">
                    {isICMP(session) ? `${session.config.payloadSize} bytes · DF ${session.config.df ? "ON" : "OFF"}` : isTCP(session) ? `Port ${session.config.port}` : isDNS(session) ? session.last?.dnsRecord || "DNS" : session.config.method}
                  </span>
                </>
              )}
            </div>
            <p title={session?.config.url}>
              {session?.config.url ?? "A clearer view of your connection"}
            </p>
            {session && (
              <div className="chart-test-meta">
                <span>Started {new Date(session.startedAt).toLocaleString()}</span>
                <span>Every {session.config.intervalMs / 1000}s</span>
                <span>{session.config.timeoutMs / 1000}s timeout</span>
                {isNetwork(session) ? <span>{isTCP(session) || isICMP(session) ? `IP ${session.config.resolvedIP || "—"} · ` : ""}Recording {session.config.recording ? "ON" : "OFF"}{session.index ? ` · Test #${session.index}` : ""}{session.last ? ` · Seq ${session.last.sequence - 1}` : ""}</span> : <><span>{session.config.followRedirects ? "Follows redirects" : "First response only"}</span>
                <span>
                  Expect {session.config.acceptedStatuses.join(", ")}
                </span>
                {session.config.proxy.enabled && <span>Via proxy</span>}<span>Recording {isRecording(session) ? "ON" : "OFF"}</span></>}
                <span>{session.running ? `Running ${duration(session)}` : `Duration ${duration(session)}`}</span>
                {session.endedAt && <span>Ended {new Date(session.endedAt).toLocaleString()}</span>}
              </div>
            )}
          </div>
          <div className="heading-actions">
            {session && session.running && !isRecording(session) && <button className="button secondary" disabled={!!busy} onClick={() => void action("record", async () => merge(await api.record(session.id)))}><Radio size={14} />Record</button>}
            {session?.running && (
              <span className="live-label">
                <i className="dot accent" />
                LIVE
              </span>
            )}
            {session && !chartID && view !== "history" && (
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
          key={session?.id ?? "empty"}
          session={session}
          onExportCSV={session && isRecording(session) && (!isNetwork(session) || !session.running) ? async () => {
            await action("export", async () => {
              const path = await api.exportCSV(session.id);
              if (path) setNotice(`Exported all saved probes to ${path}`);
            });
          } : undefined}
          onExport={session ? async (dataURL) => {
            await action("image", async () => {
              const path = await api.exportChart(session.id, dataURL);
              if (path) setNotice(`Saved chart image to ${path}`);
            });
          } : undefined}
        />
      </section>
      {view !== "history" && <section className="panel probes-panel">
        <div className="panel-heading">
          <div>
            <h2>Recent probes</h2>
            <p>The latest response, at a glance</p>
          </div>
          <span className="count-pill">
            {Math.min(6, samples.length)}
          </span>
        </div>
        {samples.length ? (
          <div className="probe-list">
            {samples
              .slice(-6)
              .reverse()
              .map((p) => (
                <div className="probe" key={p.sequence}>
                  <span className={`probe-icon ${p.success ? "ok" : "bad"}`}>
                    {p.success ? <Check size={14} /> : <X size={14} />}
                  </span>
                  <div>
                    <strong
                      className={session && isDNS(session) ? "probe-dns-result" : undefined}
                      title={session && isDNS(session) ? (p.dnsResponse || p.error || (p.success ? "Success" : "Fail")) : undefined}
                    >
                      {session && isICMP(session) ? (p.success ? "Echo reply" : p.error || "No reply") : session && isTCP(session) ? (p.success ? "Connected" : p.error || "Connection failed") : session && isDNS(session) ? (p.dnsResponse || p.error || (p.success ? "Success" : "Fail")) : p.statusCode ? `HTTP ${p.statusCode}${p.responsePhase ? ` · ${p.responsePhase}` : ""}` : "Request failed"}
                    </strong>
                    <span title={p.error || undefined}>
                      {p.error || `${clock(p.time)} · #${session && isNetwork(session) ? p.sequence - 1 : p.sequence}`}
                    </span>
                  </div>
                  <strong className="probe-rtt">
                    {session && isNetwork(session) && !p.success ? "—" : ms(p.rtt)}
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
      </section>}
    </div>
  );

  return (
    <div className={`app ${chartID ? "chart-app" : ""}`}>
      {deleting && <ActionDialog title="Delete saved test?" onClose={() => setDeleting(null)}>
        <p>{deleting.config.url}</p>
        <p>This permanently deletes the test and all its saved probes.</p>
        {error && <p role="alert">{error}</p>}
        <div className="dialog-actions">
          <button className="button secondary" onClick={() => setDeleting(null)}>Cancel</button>
          <button className="button danger" disabled={!!busy} onClick={() => void action(deleting.id, async () => {
            await api.remove(deleting.id);
            if (selected === deleting.id) setSelected(null);
            setDeleting(null); refresh();
          })}>Delete test</button>
        </div>
      </ActionDialog>}
      {!!bulkDeleting.length && <ActionDialog title={`Delete ${bulkDeleting.length} saved tests?`} onClose={() => { if (!busy) setBulkDeleting([]); }}>
        <p>This permanently deletes the selected tests and all their saved probes.</p>
        {error && <p role="alert">{error}</p>}
        <div className="dialog-actions">
          <button className="button secondary" disabled={!!busy} onClick={() => setBulkDeleting([])}>Cancel</button>
          <button className="button danger" disabled={!!busy} onClick={() => void action("bulk-delete", async () => {
            for (const id of bulkDeleting) {
              await api.remove(id);
              setBulkDeleting(current => current.filter(item => item !== id));
              setChecked(current => new Set([...current].filter(item => item !== id)));
            }
            if (selected && bulkDeleting.includes(selected)) setSelected(null);
            setChecked(new Set()); setBulkDeleting([]); refresh();
          })}>Delete {bulkDeleting.length} tests</button>
        </div>
      </ActionDialog>}
      {replaying && <ActionDialog title="Proxy password" onClose={() => { setReplaying(null); setPassword(""); }}>
        <p>Enter the proxy password to run {replaying.config.url} again. Passwords are not saved.</p>
        {error && <p role="alert">{error}</p>}
        <form onSubmit={e => { e.preventDefault(); void runAgain(replaying, password); }}>
          <label>Password<input type="password" autoComplete="off" autoFocus required value={password} onChange={e => setPassword(e.target.value)} /></label>
          <div className="dialog-actions">
            <button type="button" className="button secondary" onClick={() => { setReplaying(null); setPassword(""); }}>Cancel</button>
            <button className="button primary" disabled={!!busy}>Run again</button>
          </div>
        </form>
      </ActionDialog>}
      {aboutOpen && <ActionDialog title="About NET-Test" onClose={() => setAboutOpen(false)}>
        <div className="about-heading">
          <img src="/net-test.svg" alt="" width="64" height="64" />
          <div>
            <strong>NET-Test</strong>
            <span>Network diagnostics</span>
          </div>
        </div>
        <p className="about-description">
          NET-Test is a desktop app for checking endpoint responsiveness, spotting failures,
          and reviewing saved network test results over time.
        </p>
        <dl className="about-details">
          <div><dt>Version</dt><dd>2.0.0</dd></div>
          <div><dt>Developed by</dt><dd>Dennis Jian</dd></div>
          <div>
            <dt>Project home</dt>
            <dd>
              <button type="button" className="about-link" onClick={() => void Browser.OpenURL("https://github.com/djian01/nt_gui")}>
                github.com/djian01/nt_gui <ExternalLink size={14} />
              </button>
            </dd>
          </div>
        </dl>
        <div className="dialog-actions">
          <button type="button" className="button primary" onClick={() => setAboutOpen(false)}>Close</button>
        </div>
      </ActionDialog>}
      {!chartID && (
        <aside className="sidebar">
          <button type="button" className="brand" onClick={() => setAboutOpen(true)} aria-label="About NET-Test" title="About NET-Test">
            <img className="brand-mark" src="/net-test.svg" alt="" width="40" height="40" />
            <div>
              <strong>NET-Test</strong>
              <span>Network diagnostics</span>
            </div>
          </button>
          <div className="workspace-label">WORKSPACE</div>
          <button className={`sidebar-nav ${view === "http" ? "sidebar-current" : ""}`} onClick={() => { setView("http"); setSelected(null); setCursors([0]); }}>
            <Globe2 size={18} /><span>HTTP Tests</span><ChevronRight size={15} />
          </button>
          <button className={`sidebar-nav ${view === "icmp" ? "sidebar-current" : ""}`} onClick={() => { setView("icmp"); setSelected(null); setCursors([0]); }}>
            <Activity size={18} /><span>ICMP Tests</span><ChevronRight size={15} />
          </button>
          <button className={`sidebar-nav ${view === "tcp" ? "sidebar-current" : ""}`} onClick={() => { setView("tcp"); setSelected(null); setCursors([0]); }}>
            <Radio size={18} /><span>TCP Tests</span><ChevronRight size={15} />
          </button>
          <button className={`sidebar-nav ${view === "dns" ? "sidebar-current" : ""}`} onClick={() => { setView("dns"); setSelected(null); setCursors([0]); }}>
            <Server size={18} /><span>DNS Tests</span><ChevronRight size={15} />
          </button>
          <button className={`sidebar-nav ${view === "history" ? "sidebar-current" : ""}`} onClick={() => { setView("history"); setSelected(null); setCursors([0]); }}>
            <History size={18} /><span>History</span><ChevronRight size={15} />
          </button>
          <div className="sidebar-note">
            <Layers3 size={17} />
            <div>
              <strong>Your results, saved.</strong>
              <p>Recording off keeps tests temporary. Enable recording to save results in History. Live charts work either way.</p>
            </div>
          </div>
        </aside>
      )}
      <main>
        <header className="topbar">
          <div className="breadcrumb">
            Network tools
            <ChevronRight size={13} />
            <strong>{chartID ? "Latency detail" : view === "history" ? "Saved history" : view === "icmp" ? "ICMP Tests" : view === "tcp" ? "TCP Tests" : view === "dns" ? "DNS Tests" : "HTTP Tests"}</strong>
          </div>
          <div className="topbar-actions">
            <span className={`connection ${connected ? "" : "disconnected"}`}>
              <i className="dot" />
              {connected ? "Desktop connected" : "Connecting to desktop"}
            </span>
            <button
              type="button"
              className="theme-toggle"
              aria-label={`Switch to ${theme === "dark" ? "light" : "dark"} theme`}
              title={`Switch to ${theme === "dark" ? "light" : "dark"} theme`}
              onClick={() => setTheme((current) => current === "dark" ? "light" : "dark")}
            >
              <span className="theme-toggle-icon" aria-hidden="true">
                {theme === "dark" ? <Moon size={14} /> : <Sun size={14} />}
              </span>
              <span>{theme === "dark" ? "Dark" : "Light"}</span>
            </button>
          </div>
        </header>
        <div className="page">
          <div className="page-heading">
            <div>
              <div className="eyebrow">
                {chartID ? "TEST INSIGHTS" : "OBSERVE YOUR ENDPOINTS"}
              </div>
              <h1>
                {chartID ? "Latency detail" : view === "history" ? "Saved history" : view === "icmp" ? "ICMP Tests" : view === "tcp" ? "TCP Tests" : view === "dns" ? "DNS Tests" : "HTTP Tests"}
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
              ) : view === "history" ? (
                <button className="button secondary" disabled={!!busy || active > 0} title={active > 0 ? "Stop active tests before importing CSV results" : "Import HTTP, DNS, TCP, or ICMP results from CSV"} onClick={() => void action("import", async () => {
                  const imported = await api.importCSV();
                  if (!imported) return;
                  merge(imported); setSelected(null); setCursors([0]); refresh();
                  setNotice(`Imported ${imported.config.url}`);
                })}><Upload size={14} /> Import CSV</button>
              ) : (
                <span className="running-count">
                  <i className={`dot ${active ? "accent" : ""}`} />
                  {active} active {active === 1 ? "test" : "tests"}
                </span>
              )}
            </div>
          </div>
          {notice && <div className="notice-banner" role="status">{notice}</div>}
          {session?.importNote && <div className="notice-banner" role="note">{session.importNote}</div>}
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
          {!chartID && view === "dns" && <DNSForm disabled={!!busy || !connected || active >= 8} busy={busy === "start"} onStart={(dnsConfig, resolvers) => action("start", async () => {
            const created = await api.startDNS(dnsConfig, resolvers);
            setSelected(created[0]?.id ?? null); setCursors([0]); refresh();
          })} />}
          {!chartID && view === "icmp" && <ICMPForm disabled={!!busy || !connected || active >= 8} busy={busy === "start"} onStart={(icmpConfig, targets) => action("start", async () => {
            const created = await api.startICMP(icmpConfig, targets);
            setSelected(created[0]?.id ?? null); setCursors([0]); refresh();
          })} />}
          {!chartID && view === "tcp" && <TCPForm disabled={!!busy || !connected || active >= 8} busy={busy === "start"} onStart={(tcpConfig, targets) => action("start", async () => {
            const created = await api.startTCP(tcpConfig, targets);
            setSelected(created[0]?.id ?? null); setCursors([0]); refresh();
          })} />}
          {!chartID && view === "http" && (
            <form className="test-form panel" onSubmit={start}>
              <div className="form-heading">
                <span className="section-icon">
                  <Plus size={16} />
                </span>
                <h2>New test</h2>
                <label className="recording-option"><input type="checkbox" checked={!!config.recording} onChange={e => setConfig({ ...config, recording: e.target.checked })} />Result Recording {config.recording ? "ON" : "OFF"}</label>
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
                    <option>POST</option>
                    <option>PATCH</option>
                  </select>
                </label>
                <label>
                  Interval
                  <input
                    type="number"
                    min={1}
                    step={1}
                    value={config.intervalMs / 1000}
                    onChange={(e) =>
                      setConfig({
                        ...config,
                        intervalMs: Number(e.target.value) * 1000,
                      })
                    }
                  />
                </label>
                <label>
                  Timeout
                  <input
                    type="number"
                    min={1}
                    step={1}
                    value={config.timeoutMs / 1000}
                    onChange={(e) =>
                      setConfig({
                        ...config,
                        timeoutMs: Number(e.target.value) * 1000,
                      })
                    }
                  />
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
                  <div className="advanced-section redirect-options">
                    <div className="advanced-copy">
                      <span className="advanced-icon">
                        <ArrowUpRight size={16} />
                      </span>
                      <div>
                        <strong>HTTP redirects</strong>
                        <p>Follow redirects and measure the final response.</p>
                      </div>
                    </div>
                    <label className="switch-control">
                      <input
                        type="checkbox"
                        role="switch"
                        checked={!!config.followRedirects}
                        onChange={(event) => setConfig({ ...config, followRedirects: event.target.checked })}
                      />
                      <span className="switch-track"><i /></span>
                      {config.followRedirects ? "Follow" : "First response"}
                    </label>
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
          {!chartID && (
            <section className="panel sessions-panel">
              <div className="panel-heading">
                <div className="title-with-count">
                  <h2>{view === "history" ? "Saved tests" : "Session tests"}</h2>
                  <span className="count-pill">{filtered.length}</span>
                </div>
                <div className="table-tools">
                  {view === "history" && checked.size > 0 && (
                    <button className="reset-zoom bulk-delete" disabled={!!busy} onClick={() => setBulkDeleting([...checked])}>
                      <Trash2 size={12} /> Delete selected ({checked.size})
                    </button>
                  )}
                  {view === "history" && (
                    <select className="history-type-filter" aria-label="Filter history by test type"
                      value={historyType} onChange={(event) => {
                        setHistoryType(event.target.value);
                        setCursors([0]);
                        setSelected(null);
                        setChecked(new Set());
                      }}>
                      <option value="all">All types</option>
                      <option value="http">HTTP</option>
                      <option value="dns">DNS</option>
                      <option value="tcp">TCP</option>
                      <option value="icmp">ICMP</option>
                    </select>
                  )}
                  <label className="search">
                    <Search size={14} />
                    <input
                      aria-label="Filter tests by endpoint"
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
                      {view === "history" && <th className="select-cell">
                        <input
                          type="checkbox"
                          aria-label="Select all saved tests on this page"
                          checked={filtered.length > 0 && filtered.every((s) => checked.has(s.id))}
                          onChange={(event) => setChecked(event.target.checked ? new Set(filtered.map((s) => s.id)) : new Set())}
                        />
                      </th>}
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
                        onClick={() => selectTest(s.id)}
                        onKeyDown={(event) => {
                          if (
                            event.target === event.currentTarget &&
                            (event.key === "Enter" || event.key === " ")
                          ) {
                            event.preventDefault();
                            selectTest(s.id);
                          }
                        }}
                      >
                        {view === "history" && <td className="select-cell" onClick={(event) => event.stopPropagation()}>
                          <input
                            type="checkbox"
                            aria-label={`Select saved test ${s.config.url}`}
                            checked={checked.has(s.id)}
                            onChange={(event) => setChecked((current) => {
                              const next = new Set(current);
                              if (event.target.checked) next.add(s.id); else next.delete(s.id);
                              return next;
                            })}
                          />
                        </td>}
                        <td>
                          <div className="endpoint-button">
                            <span className="endpoint-icon">
                              {isICMP(s) ? <Activity size={17} /> : isTCP(s) ? <Radio size={17} /> : isDNS(s) ? <Server size={17} /> : <Globe2 size={17} />}
                            </span>
                            <span>
                              <strong title={s.config.url}>
                                {s.config.url}
                              </strong>
                              <small>
                                {isICMP(s) ? `ICMP / ${s.config.payloadSize} bytes / DF ${s.config.df ? "ON" : "OFF"}` : isTCP(s) ? `TCP / Port ${s.config.port}` : isDNS(s) ? `DNS / ${s.config.protocol?.toUpperCase()}` : s.config.method} <span>·</span> Started{" "}
                                {new Date(s.startedAt).toLocaleString()}
                              </small>
                              <small className="endpoint-detail">Recording {isRecording(s) ? "ON" : "OFF"}</small>
                            </span>
                          </div>
                        </td>
                        <td>
                          <Badge session={s} />
                          {(isTCP(s) || isICMP(s)) && <span className="dns-result" title={`Resolved IP: ${s.config.resolvedIP || "—"}`} aria-label={`Resolved IP: ${s.config.resolvedIP || "—"}`}>{s.config.resolvedIP || "—"}{s.index ? ` · Test #${s.index}` : ""}{s.last ? ` · Seq ${s.last.sequence - 1}` : ""}</span>}
                          {isDNS(s) && <span className="dns-result" title={dnsResult(s)} aria-label={`DNS response: ${dnsResult(s)}`}>
                            {shortDNSResult(s)}
                          </span>}
                        </td>
                        <td className="numeric">
                          {s.last && (!isNetwork(s) || s.last.success) ? `${ms(s.last.rtt)} ms` : "—"}
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
                            {view !== "history" && <button
                              className="icon-button"
                              title="Open chart"
                              aria-label={`Open chart for ${s.config.url}`}
                              disabled={!!busy || (isNetwork(s) && s.sent < 3)}
                              onClick={() =>
                                void action("chart", () => api.chart(s.id))
                              }
                            >
                              <ExternalLink size={15} />
                            </button>}
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
                                  ? "Stop test first"
                                  : !isRecording(s) ? "Close temporary test" : (view === "dns" || view === "tcp" || view === "icmp") ? "Close row (keep history)" : "Delete saved test"
                              }
                              aria-label={`${view === "dns" || view === "tcp" || view === "icmp" || !isRecording(s) ? "Close row for" : "Delete saved test"} ${s.config.url}`}
                              disabled={s.running || !!busy}
                              onClick={() => {
                                if (view === "dns" || view === "tcp" || view === "icmp") {
                                  void action(s.id, async () => {
                                    await api.dismiss(s.id);
                                    if (selected === s.id) setSelected(null);
                                    refresh();
                                  });
                                } else if (!isRecording(s)) {
                                  void action(s.id, async () => {
                                    await api.remove(s.id);
                                    if (selected === s.id) setSelected(null);
                                    refresh();
                                  });
                                } else setDeleting(s);
                              }}
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
                    {view === "icmp" ? <Activity size={21} /> : view === "tcp" ? <Radio size={21} /> : view === "dns" ? <Server size={21} /> : view === "history" ? <History size={21} /> : <Globe2 size={21} />}
                  </span>
                  <div>
                    <strong>
                      {search
                        ? "No matching endpoints"
                        : view === "history" ? "No saved tests yet." : "No tests this session."}
                    </strong>
                    <p>
                      {search
                        ? "Try another search term."
                        : view === "history"
                          ? "Stopped tests and imported results will appear here."
                          : view === "icmp" ? "Enter IPv4 addresses or hostnames above to start ICMP tests." : view === "tcp" ? "Enter TCP server IPs or hostnames and a port above to start." : view === "dns" ? "Enter resolver IPs and a DNS query above to start." : "Enter a URL above and start a test. Results stay here until deleted or the app closes."}
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
                <div className="pagination">
                  <button className="reset-zoom" disabled={loading || cursors.length === 1} onClick={() => setCursors(v => v.slice(0, -1))}>Previous</button>
                  <span>{loading ? "Loading…" : `Page ${cursors.length} · ${filtered.length} tests`}</span>
                  <button className="reset-zoom" disabled={loading || !next} onClick={() => setCursors(v => [...v, next])}>Next</button>
                </div>
              </div>
            </section>
          )}
          {(view !== "history" || session) && metrics}
          {(view !== "history" || session) && analysis}
          <footer className="page-footer">
            <span>
              <ShieldCheck size={14} />
              {chartID
                ? "Live session · shared with your main window"
                : "Runs locally on your desktop"}
            </span>
            <span>
              Probes saved when recording is ON · live charts always available
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
      initial={false}
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

function ActionDialog({ title, onClose, children }: { title: string; onClose: () => void; children: ReactNode }) {
  const ref = useRef<HTMLDialogElement>(null);
  useEffect(() => { ref.current?.showModal(); }, []);
  return <dialog ref={ref} className="action-dialog" aria-label={title} onCancel={e => { e.preventDefault(); onClose(); }}>
    <h2>{title}</h2>{children}
  </dialog>;
}
