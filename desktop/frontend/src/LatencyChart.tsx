import { useEffect, useRef, useState, type CSSProperties } from "react";
import { Activity, Download, RotateCcw, Pause, Play, ImageDown } from "lucide-react";
import { api, errorText, type Sample, type Session, type Timeline } from "./api";

export const ms = (value: number) =>
  value.toLocaleString(undefined, {
    maximumFractionDigits: 1,
    minimumFractionDigits: 1,
  });
export const clock = (value: string) =>
  new Date(value).toLocaleTimeString(undefined, { hour12: false });

type PlotPoint = Sample & { plotX: number; plotY: number };
type TimelineWindow = { start: number | null; end: number | null };

function smoothPath(points: PlotPoint[]) {
  if (!points.length) return "";
  if (points.length === 1) return `M ${points[0].plotX} ${points[0].plotY}`;
  return points.slice(1).reduce((path, point, index) => {
    const previous = points[index];
    const midpoint = (previous.plotX + point.plotX) / 2;
    return `${path} C ${midpoint} ${previous.plotY}, ${midpoint} ${point.plotY}, ${point.plotX} ${point.plotY}`;
  }, `M ${points[0].plotX} ${points[0].plotY}`);
}

function nearestByTime(samples: Sample[], target: number) {
  let low = 0;
  let high = samples.length - 1;
  while (low < high) {
    const middle = Math.floor((low + high) / 2);
    if (Date.parse(samples[middle].time) < target) low = middle + 1;
    else high = middle;
  }
  if (
    low > 0 &&
    Math.abs(Date.parse(samples[low - 1].time) - target) <
      Math.abs(Date.parse(samples[low].time) - target)
  )
    return low - 1;
  return low;
}

async function chartPNG(svg: SVGSVGElement, session: Session, timeline: Timeline) {
  const clone = svg.cloneNode(true) as SVGSVGElement;
  clone.setAttribute("width", "1680");
  clone.setAttribute("height", "410");
  const styles = getComputedStyle(svg);
  const variables = ["--accent", "--accent-text", "--border", "--text-muted", "--surface", "--danger"]
    .map((name) => `${name}:${styles.getPropertyValue(name).trim()}`)
    .join(";");
  const style = document.createElementNS("http://www.w3.org/2000/svg", "style");
  style.textContent = `:root{${variables}}`;
  clone.prepend(style);

  const source = new Blob([new XMLSerializer().serializeToString(clone)], {
    type: "image/svg+xml;charset=utf-8",
  });
  const objectURL = URL.createObjectURL(source);
  try {
    const image = new Image();
    await new Promise<void>((resolve, reject) => {
      image.onload = () => resolve();
      image.onerror = () => reject(new Error("Could not render the chart image"));
      image.src = objectURL;
    });
    const canvas = document.createElement("canvas");
    canvas.width = 1800;
    canvas.height = 690;
    const context = canvas.getContext("2d");
    if (!context) throw new Error("Chart image export is unavailable");
    const background = styles.getPropertyValue("--surface").trim() || "#111d30";
    const foreground = styles.getPropertyValue("--text").trim() || "#e4ecf7";
    const muted = styles.getPropertyValue("--text-muted").trim() || "#96a9c3";
    context.fillStyle = background;
    context.fillRect(0, 0, canvas.width, canvas.height);
    context.fillStyle = foreground;
    context.font = "600 34px Inter, system-ui, sans-serif";
    context.fillText(session.config.type === "icmp" ? "Net Test · ICMP echo time" : session.config.type === "tcp" ? "Net Test · TCP connection time" : session.config.type === "dns" ? "Net Test · DNS response time" : "Net Test · HTTP response time", 60, 64);
    context.fillStyle = muted;
    context.font = "24px Inter, system-ui, sans-serif";
    const endpoint = session.config.url.length > 100 ? `${session.config.url.slice(0, 97)}…` : session.config.url;
    context.fillText(endpoint, 60, 105);
    context.font = "20px Inter, system-ui, sans-serif";
    context.fillText(
      `${session.config.type === "icmp" ? `ICMP / ${session.config.payloadSize} bytes / DF ${session.config.df ? "ON" : "OFF"}` : session.config.type === "tcp" ? `TCP / Port ${session.config.port}` : session.config.type === "dns" ? `DNS / ${session.config.protocol?.toUpperCase()}` : session.config.method} · ${timeline.count.toLocaleString()} probes · ${timeline.succeeded.toLocaleString()} accepted · Avg ${ms(timeline.average)} ms · Max ${ms(timeline.maximum)} ms`,
      60,
      140,
    );
    context.drawImage(image, 60, 175, 1680, 410);
    const axisMaximum = Math.max(10, timeline.maximum) * 1.15;
    context.textAlign = "right";
    context.font = "16px Inter, system-ui, sans-serif";
    [1, 0.75, 0.5, 0.25, 0].forEach((fraction) => {
      context.fillText(String(Math.round(axisMaximum * fraction)), 52, 175 + (207 - fraction * 192) * 410 / 220 + 6);
    });
    context.textAlign = "left";
    context.fillText(clock(new Date(timeline.from).toISOString()), 75, 602);
    context.textAlign = "center";
    context.fillText(clock(new Date((timeline.from + timeline.to) / 2).toISOString()), 900, 602);
    context.textAlign = "right";
    context.fillText(clock(new Date(timeline.to).toISOString()), 1725, 602);
    context.textAlign = "left";
    context.font = "20px Inter, system-ui, sans-serif";
    context.fillText(
      `${new Date(timeline.from).toLocaleString()} – ${new Date(timeline.to).toLocaleString()}`,
      60,
      650,
    );
    return canvas.toDataURL("image/png");
  } finally {
    URL.revokeObjectURL(objectURL);
  }
}

export default function LatencyChart({
  session,
  onExportCSV,
  onExport,
}: {
  session: Session | null;
  onExportCSV?: () => Promise<void>;
  onExport?: (dataURL: string) => Promise<void>;
}) {
  const [window, setWindow] = useState<TimelineWindow>({ start: null, end: null });
  const [hover, setHover] = useState<number | null>(null);
  const [now, setNow] = useState(Date.now());
  const [pausedAt, setPausedAt] = useState<number | null>(null);
  const [timeline, setTimeline] = useState<Timeline | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [exportingCSV, setExportingCSV] = useState(false);
  const [exporting, setExporting] = useState(false);
  const svgRef = useRef<SVGSVGElement>(null);
  const origin = session ? Date.parse(session.startedAt) : now;
  const domainEnd = Math.max(origin + 1, pausedAt ?? (session?.endedAt ? Date.parse(session.endedAt) : now));
  const selectedStart = Math.max(origin, Math.min(window.start ?? origin, domainEnd - 1));
  const selectedEnd = Math.max(selectedStart + 1, Math.min(window.end ?? domainEnd, domainEnd));
  const revision = pausedAt === null ? session?.revision : 0;
  const id = session?.id;
  useEffect(() => {
    if (!session?.running || pausedAt !== null) return;
    const timer = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(timer);
  }, [session?.running, pausedAt]);
  useEffect(() => {
    if (session && !session.running) setPausedAt(null);
  }, [session?.running]);
  useEffect(() => {
    if (!id) return;
    let current = true;
    const timer = setTimeout(() => {
      setLoading(true);
      api.timeline(id, selectedStart, selectedEnd).then(value => {
        if (current) { setTimeline(value); setError(""); }
      }).catch(err => { if (current) setError(errorText(err)); })
        .finally(() => { if (current) setLoading(false); });
    }, 100);
    return () => { current = false; clearTimeout(timer); };
  }, [id, selectedStart, selectedEnd, revision]);
  const points = timeline?.samples ?? [];
  const plotted = points;
  const maxRTT = Math.max(10, timeline?.maximum ?? 0) * 1.15;
  const start = timeline?.from ?? selectedStart;
  const end = Math.max(start + 1, timeline?.to ?? selectedEnd);
  const x = (point: Sample) =>
    8 + ((Date.parse(point.time) - start) / Math.max(1, end - start)) * 884;
  const y = (point: Sample) =>
    207 - ((point.success ? point.rtt : 0) / maxRTT) * 192;
  const segments = plotted
    .reduce<PlotPoint[][]>(
      (result, point) => {
        if (!point.success) return [...result, []];
        result.at(-1)!.push({ ...point, plotX: x(point), plotY: y(point) });
        return result;
      },
      [[]],
    )
    .filter((segment) => segment.length);
  const path = segments.map(smoothPath).join(" ");
  const areaPath = segments
    .map(
      (segment) =>
        `${smoothPath(segment)} L ${segment.at(-1)!.plotX} 207 L ${segment[0].plotX} 207 Z`,
    )
    .join(" ");
  const average = timeline?.average ?? 0;
  const latest = points.at(-1);
  const hovered = hover == null ? null : points[Math.min(hover, points.length - 1)];
  const zoomed = window.start !== null || window.end !== null;
  const rangeStyle = {
    "--range-start": `${((selectedStart - origin) / (domainEnd - origin)) * 100}%`,
    "--range-end": `${((selectedEnd - origin) / (domainEnd - origin)) * 100}%`,
  } as CSSProperties;

  function resetZoom() {
    setWindow({ start: null, end: null });
    setPausedAt(null); setNow(Date.now());
    setHover(null);
  }

  return (
    <>
      {error && <div className="error-banner" role="alert">{error}</div>}
      <div className="chart-toolbar">
        <span>
          <i className="legend-line" /> Response time <small>milliseconds</small>
        </span>
        <span className="visible-window">
          {points.length
            ? `${clock(new Date(start).toISOString())} – ${clock(new Date(end).toISOString())}`
            : "Complete timeline"}
        </span>
        <div className="chart-actions">
          {session && onExportCSV && (
            <button className="reset-zoom" disabled={exportingCSV} onClick={() => {
              setExportingCSV(true);
              void onExportCSV().finally(() => setExportingCSV(false));
            }}>
              <Download size={12} /> {exportingCSV ? "Exporting…" : "Export CSV"}
            </button>
          )}
          {session?.running && (
            <button className="reset-zoom" onClick={() => { setPausedAt(pausedAt === null ? domainEnd : null); setNow(Date.now()); }}>
              {pausedAt === null ? <Pause size={12} /> : <Play size={12} />}
              {pausedAt === null ? "Pause chart" : "Resume chart"}
            </button>
          )}
          {session && !session.running && timeline && onExport && (
            <button className="reset-zoom" disabled={exporting || !points.length} onClick={() => {
              if (!svgRef.current) return;
              setExporting(true); setError("");
              void chartPNG(svgRef.current, session, timeline)
                .then(onExport)
                .catch((err) => setError(errorText(err)))
                .finally(() => setExporting(false));
            }}>
              <ImageDown size={12} /> {exporting ? "Preparing…" : "Save PNG"}
            </button>
          )}
        </div>
      </div>
      {!points.length ? (
        <div className="chart-empty">
          <div className="empty-pulse">
            <Activity size={28} />
          </div>
          <strong>{loading ? "Loading saved results…" : "No probes in this period"}</strong>
          <p>{session ? "Choose a wider time range or wait for a completed probe." : "Start a test to see response times as they arrive."}</p>
          <div className="empty-grid" />
        </div>
      ) : (
        <>
          <div className="plot">
            <div className="y-labels">
              {[1, 0.75, 0.5, 0.25, 0].map((value) => (
                <span key={value}>{Math.round(maxRTT * value)}</span>
              ))}
            </div>
            <div className="plot-area">
              <svg
                ref={svgRef}
                viewBox="0 0 900 220"
                preserveAspectRatio="none"
                role="img"
                aria-label={`Latency chart showing ${points.length} probes. Latest: ${latest?.success ? ms(latest.rtt) + " milliseconds" : "failed"}.`}
                onPointerLeave={() => setHover(null)}
                onPointerMove={(event) => {
                  const rect = event.currentTarget.getBoundingClientRect();
                  const ratio = Math.min(
                    1,
                    Math.max(0, (event.clientX - rect.left) / rect.width),
                  );
                  setHover(nearestByTime(points, start + ratio * (end - start)));
                }}
              >
                <defs>
                  <linearGradient id="latency-area" x1="0" x2="0" y1="0" y2="1">
                    <stop offset="0%" stopColor="var(--accent)" stopOpacity="0.34" />
                    <stop offset="65%" stopColor="var(--accent)" stopOpacity="0.08" />
                    <stop offset="100%" stopColor="var(--accent)" stopOpacity="0" />
                  </linearGradient>
                  <filter id="line-glow" x="-10%" y="-30%" width="120%" height="160%">
                    <feGaussianBlur stdDeviation="3" result="blur" />
                    <feMerge>
                      <feMergeNode in="blur" />
                      <feMergeNode in="SourceGraphic" />
                    </feMerge>
                  </filter>
                </defs>
                {[15, 63, 111, 159, 207].map((value) => (
                  <line
                    key={value}
                    x1="0"
                    x2="900"
                    y1={value}
                    y2={value}
                    stroke="var(--border)"
                    strokeDasharray="2 7"
                    vectorEffect="non-scaling-stroke"
                  />
                ))}
                {[8, 229, 450, 671, 892].map((value) => (
                  <line
                    key={`vertical-${value}`}
                    x1={value}
                    x2={value}
                    y1="15"
                    y2="207"
                    stroke="var(--border)"
                    strokeOpacity="0.42"
                    vectorEffect="non-scaling-stroke"
                  />
                ))}
                {average > 0 && (
                  <line
                    x1="8"
                    x2="892"
                    y1={207 - (average / maxRTT) * 192}
                    y2={207 - (average / maxRTT) * 192}
                    stroke="var(--accent-text)"
                    strokeOpacity="0.45"
                    strokeDasharray="5 6"
                    vectorEffect="non-scaling-stroke"
                  />
                )}
                <path d={areaPath} fill="url(#latency-area)" />
                <path
                  d={path}
                  fill="none"
                  stroke="var(--accent)"
                  strokeWidth="3"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  filter="url(#line-glow)"
                  vectorEffect="non-scaling-stroke"
                />
                {plotted
                  .filter((point) => !point.success || plotted.length < 3)
                  .map((point) => (
                    <circle
                      key={point.sequence}
                      cx={x(point)}
                      cy={y(point)}
                      r="3.5"
                      fill={point.success ? "var(--accent)" : "var(--danger)"}
                    />
                  ))}
                {hovered && (
                  <>
                    <line
                      x1={x(hovered)}
                      x2={x(hovered)}
                      y1="0"
                      y2="220"
                      stroke="var(--text-muted)"
                      strokeDasharray="4 4"
                    />
                    <circle
                      cx={x(hovered)}
                      cy={y(hovered)}
                      r="5"
                      fill={hovered.success ? "var(--accent)" : "var(--danger)"}
                      stroke="var(--surface)"
                      strokeWidth="2"
                    />
                  </>
                )}
                {!hovered && latest?.success && (
                  <>
                    <circle
                      className="latest-pulse"
                      cx={x(latest)}
                      cy={y(latest)}
                      r="9"
                      fill="var(--accent)"
                    />
                    <circle
                      cx={x(latest)}
                      cy={y(latest)}
                      r="4"
                      fill="var(--accent)"
                      stroke="var(--surface)"
                      strokeWidth="2"
                    />
                  </>
                )}
              </svg>
              <div className="chart-summary">
                <span>AVG</span>
                <strong>{average ? ms(average) : "—"}</strong>
                <small>ms</small>
              </div>
              {hovered && (
                <div
                  className="chart-tooltip"
                  style={{ left: `${Math.min(76, Math.max(3, x(hovered) / 9))}%` }}
                >
                  <span>
                    PROBE #{session?.config.type === "dns" || session?.config.type === "tcp" || session?.config.type === "icmp" ? hovered.sequence - 1 : hovered.sequence} · {clock(hovered.time)}
                  </span>
                  <strong>
                    {hovered.success
                      ? `${ms(hovered.rtt)} ms`
                      : hovered.error || "Failed"}
                  </strong>
                  <span>
                    {session?.config.type === "icmp" ? `ICMP · ${hovered.success ? "Echo reply" : hovered.error || "No reply"}` : session?.config.type === "tcp" ? `TCP · ${hovered.success ? "Connected" : hovered.error || "Connection failed"}` : session?.config.type === "dns" ? `${hovered.dnsRecord || "DNS"} · ${hovered.dnsResponse || hovered.error || "No response"}` : hovered.statusCode
                      ? `HTTP ${hovered.statusCode}${hovered.responsePhase ? ` · ${hovered.responsePhase}` : ""}`
                      : "No response"}
                  </span>
                </div>
              )}
              <div className="x-labels">
                <span>{clock(new Date(start).toISOString())}</span>
                <span>{clock(new Date((start + end) / 2).toISOString())}</span>
                <span>{clock(new Date(end).toISOString())}</span>
              </div>
            </div>
          </div>

        </>
      )}
      {session && (
          <div className="timeline-navigator">
            <div className="timeline-heading">
              <div>
                <strong>Timeline zoom</strong>
                <span>Drag either handle to choose the visible period.</span>
              </div>
              <button
                type="button"
                className="reset-zoom"
                disabled={!zoomed && pausedAt === null}
                onClick={resetZoom}
              >
                <RotateCcw size={12} /> Reset zoom
              </button>
            </div>
            <div className="dual-range" style={rangeStyle}>
              <div className="range-track">
                <span />
              </div>
              <input
                className="range-start"
                type="range"
                min={origin}
                max={domainEnd}
                step={1}
                value={selectedStart}
                disabled={!session || session.sent < 2}
                aria-label="Visible timeline start"
                onChange={(event) => {
                  setWindow({
                    start: Math.min(Number(event.target.value), selectedEnd - 1),
                    end: selectedEnd,
                  });
                  setHover(null);
                }}
              />
              <input
                className="range-end"
                type="range"
                min={origin}
                max={domainEnd}
                step={1}
                value={selectedEnd}
                disabled={!session || session.sent < 2}
                aria-label="Visible timeline end"
                onChange={(event) => {
                  const next = Math.max(Number(event.target.value), selectedStart + 1);
                  setWindow({
                    start: selectedStart,
                    end: next,
                  });
                  setHover(null);
                }}
              />
            </div>
            <div className="timeline-labels">
              <span>Start · {clock(new Date(origin).toISOString())}</span>
              <strong>
                {(timeline?.count ?? 0).toLocaleString()} of {(session?.sent ?? 0).toLocaleString()} probes
              </strong>
              <span>Latest · {clock(new Date(domainEnd).toISOString())}</span>
            </div>
          </div>
      )}
      <div className="chart-footnote">
        <span>
          <i className="dot accent" />
          Successful response <i className="dot red" />
          Failed probe
        </span>
        <span>
          {pausedAt !== null ? "Chart paused · saving continues" : timeline?.aggregated ? "Overview · zoom in for individual probes" : "Individual probes"}
        </span>
      </div>
    </>
  );
}
