import { useEffect, useState, type CSSProperties } from "react";
import { Activity, RotateCcw, Pause, Play } from "lucide-react";
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

export default function LatencyChart({ session }: { session: Session | null }) {
  const [window, setWindow] = useState<TimelineWindow>({ start: null, end: null });
  const [hover, setHover] = useState<number | null>(null);
  const [now, setNow] = useState(Date.now());
  const [pausedAt, setPausedAt] = useState<number | null>(null);
  const [timeline, setTimeline] = useState<Timeline | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
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
        <button className="reset-zoom" disabled={!session} onClick={() => { setPausedAt(pausedAt === null ? domainEnd : null); setNow(Date.now()); }}>
          {pausedAt === null ? <Pause size={12} /> : <Play size={12} />}
          {pausedAt === null ? "Pause chart" : "Resume chart"}
        </button>
      </div>
      {!points.length ? (
        <div className="chart-empty">
          <div className="empty-pulse">
            <Activity size={28} />
          </div>
          <strong>{loading ? "Loading saved results…" : "No probes in this period"}</strong>
          <p>{session ? "Choose a wider time range or wait for a completed probe." : "Start an HTTP test to see response times as they arrive."}</p>
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
                    PROBE #{hovered.sequence} · {clock(hovered.time)}
                  </span>
                  <strong>
                    {hovered.success
                      ? `${ms(hovered.rtt)} ms`
                      : hovered.error || "Failed"}
                  </strong>
                  <span>
                    {hovered.statusCode
                      ? `HTTP ${hovered.statusCode}`
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
