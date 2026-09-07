import { useState, type CSSProperties } from "react";
import { Activity, RotateCcw } from "lucide-react";
import type { Sample } from "./api";

export const ms = (value: number) =>
  value.toLocaleString(undefined, {
    maximumFractionDigits: 1,
    minimumFractionDigits: 1,
  });
export const clock = (value: string) =>
  new Date(value).toLocaleTimeString(undefined, { hour12: false });

type PlotPoint = Sample & { plotX: number; plotY: number };
type TimelineWindow = { start: number; end: number | null };

function smoothPath(points: PlotPoint[]) {
  if (!points.length) return "";
  if (points.length === 1) return `M ${points[0].plotX} ${points[0].plotY}`;
  return points.slice(1).reduce((path, point, index) => {
    const previous = points[index];
    const midpoint = (previous.plotX + point.plotX) / 2;
    return `${path} C ${midpoint} ${previous.plotY}, ${midpoint} ${point.plotY}, ${point.plotX} ${point.plotY}`;
  }, `M ${points[0].plotX} ${points[0].plotY}`);
}

// Bound SVG work without deleting samples. Each bucket retains a failure when
// present, otherwise its highest-latency point.
function downsample(samples: Sample[], limit = 700) {
  if (samples.length <= limit) return samples;
  const result = [samples[0]];
  const buckets = limit - 2;
  for (let bucket = 0; bucket < buckets; bucket++) {
    const from = 1 + Math.floor((bucket * (samples.length - 2)) / buckets);
    const to =
      1 + Math.floor(((bucket + 1) * (samples.length - 2)) / buckets);
    let representative = samples[from];
    for (let index = from; index < Math.max(from + 1, to); index++) {
      const point = samples[index];
      if (!point) continue;
      if (!point.success) {
        representative = point;
        break;
      }
      if (!representative || point.rtt > representative.rtt)
        representative = point;
    }
    if (representative?.sequence !== result.at(-1)?.sequence)
      result.push(representative);
  }
  const last = samples.at(-1)!;
  if (last.sequence !== result.at(-1)?.sequence) result.push(last);
  return result;
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

export default function LatencyChart({ samples }: { samples: Sample[] }) {
  const [window, setWindow] = useState<TimelineWindow>({ start: 0, end: null });
  const [hover, setHover] = useState<number | null>(null);
  const lastIndex = Math.max(0, samples.length - 1);
  const selectedStart = Math.min(window.start, lastIndex);
  const selectedEnd = Math.max(
    selectedStart,
    Math.min(window.end ?? lastIndex, lastIndex),
  );
  const points =
    selectedStart === 0 && selectedEnd === lastIndex
      ? samples
      : samples.slice(selectedStart, selectedEnd + 1);
  const plotted = downsample(points);
  let maxSuccessfulRTT = 10;
  let successfulTotal = 0;
  let successfulCount = 0;
  for (const point of points) {
    if (!point.success) continue;
    maxSuccessfulRTT = Math.max(maxSuccessfulRTT, point.rtt);
    successfulTotal += point.rtt;
    successfulCount++;
  }
  const maxRTT = maxSuccessfulRTT * 1.15;
  const start = points.length ? Date.parse(points[0].time) : 0;
  const end =
    points.length > 1
      ? Date.parse(points[points.length - 1].time)
      : start + 1000;
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
  const average = successfulCount ? successfulTotal / successfulCount : 0;
  const latest = points.at(-1);
  const hovered = hover == null ? null : points[Math.min(hover, points.length - 1)];
  const zoomed = selectedStart > 0 || window.end !== null;
  const rangeStyle = {
    "--range-start": `${lastIndex ? (selectedStart / lastIndex) * 100 : 0}%`,
    "--range-end": `${lastIndex ? (selectedEnd / lastIndex) * 100 : 100}%`,
  } as CSSProperties;

  function resetZoom() {
    setWindow({ start: 0, end: null });
    setHover(null);
  }

  return (
    <>
      <div className="chart-toolbar">
        <span>
          <i className="legend-line" /> Response time <small>milliseconds</small>
        </span>
        <span className="visible-window">
          {points.length
            ? `${clock(points[0].time)} – ${clock(points.at(-1)!.time)}`
            : "Complete timeline"}
        </span>
      </div>
      {!points.length ? (
        <div className="chart-empty">
          <div className="empty-pulse">
            <Activity size={28} />
          </div>
          <strong>Your next signal starts here</strong>
          <p>Start an HTTP test to see response times as they arrive.</p>
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
                <span>{clock(points[0].time)}</span>
                <span>{clock(points[Math.floor((points.length - 1) / 2)].time)}</span>
                <span>{clock(points.at(-1)!.time)}</span>
              </div>
            </div>
          </div>
          <div className="timeline-navigator">
            <div className="timeline-heading">
              <div>
                <strong>Timeline zoom</strong>
                <span>Drag either handle to choose the visible period.</span>
              </div>
              <button
                type="button"
                className="reset-zoom"
                disabled={!zoomed}
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
                min={0}
                max={lastIndex}
                value={selectedStart}
                disabled={samples.length < 2}
                aria-label="Visible timeline start"
                onChange={(event) => {
                  setWindow({
                    start: Math.min(Number(event.target.value), selectedEnd - 1),
                    end: window.end,
                  });
                  setHover(null);
                }}
              />
              <input
                className="range-end"
                type="range"
                min={0}
                max={lastIndex}
                value={selectedEnd}
                disabled={samples.length < 2}
                aria-label="Visible timeline end"
                onChange={(event) => {
                  const next = Math.max(Number(event.target.value), selectedStart + 1);
                  setWindow({
                    start: selectedStart,
                    end: next === lastIndex ? null : next,
                  });
                  setHover(null);
                }}
              />
            </div>
            <div className="timeline-labels">
              <span>Start · {clock(samples[0].time)}</span>
              <strong>
                {points.length.toLocaleString()} of {samples.length.toLocaleString()} probes
              </strong>
              <span>Latest · {clock(samples.at(-1)!.time)}</span>
            </div>
          </div>
        </>
      )}
      <div className="chart-footnote">
        <span>
          <i className="dot accent" />
          Successful response <i className="dot red" />
          Failed probe
        </span>
        <span>
          {points.length.toLocaleString()} visible · {samples.length.toLocaleString()} total
        </span>
      </div>
    </>
  );
}
