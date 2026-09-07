import { useState } from "react";
import { Activity } from "lucide-react";
import type { Sample } from "./api";

export const ms = (value: number) =>
  value.toLocaleString(undefined, {
    maximumFractionDigits: 1,
    minimumFractionDigits: 1,
  });
export const clock = (value: string) =>
  new Date(value).toLocaleTimeString(undefined, { hour12: false });

type PlotPoint = Sample & { plotX: number; plotY: number };

function smoothPath(points: PlotPoint[]) {
  if (!points.length) return "";
  if (points.length === 1) return `M ${points[0].plotX} ${points[0].plotY}`;

  return points.slice(1).reduce((path, point, index) => {
    const previous = points[index];
    const midpoint = (previous.plotX + point.plotX) / 2;
    return `${path} C ${midpoint} ${previous.plotY}, ${midpoint} ${point.plotY}, ${point.plotX} ${point.plotY}`;
  }, `M ${points[0].plotX} ${points[0].plotY}`);
}

export default function LatencyChart({ samples }: { samples: Sample[] }) {
  const [range, setRange] = useState(120);
  const [hover, setHover] = useState<number | null>(null);
  const points = samples.slice(-range);
  const maxRTT =
    Math.max(10, ...points.filter((p) => p.success).map((p) => p.rtt)) * 1.15;
  const start = points.length ? Date.parse(points[0].time) : 0;
  const end =
    points.length > 1
      ? Date.parse(points[points.length - 1].time)
      : start + 1000;
  const x = (p: Sample) =>
    8 + ((Date.parse(p.time) - start) / Math.max(1, end - start)) * 884;
  const y = (p: Sample) => 207 - ((p.success ? p.rtt : 0) / maxRTT) * 192;
  const successful = points.filter((point) => point.success);
  const segments = points
    .reduce<PlotPoint[][]>(
      (result, point) => {
        if (!point.success) return [...result, []];
        const current = result.at(-1)!;
        current.push({ ...point, plotX: x(point), plotY: y(point) });
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
  const average = successful.length
    ? successful.reduce((sum, point) => sum + point.rtt, 0) / successful.length
    : 0;
  const latest = points.at(-1);
  const hovered =
    hover == null ? null : points[Math.min(hover, points.length - 1)];
  return (
    <>
      <div className="chart-toolbar">
        <span>
          <i className="legend-line" /> Response time{" "}
          <small>milliseconds</small>
        </span>
        <div className="segmented" aria-label="Chart range">
          {[30, 120, 600].map((n) => (
            <button
              key={n}
              aria-pressed={range === n}
              onClick={() => {
                setRange(n);
                setHover(null);
              }}
            >
              Last {n}
            </button>
          ))}
        </div>
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
        <div className="plot">
          <div className="y-labels">
            {[1, 0.75, 0.5, 0.25, 0].map((n) => (
              <span key={n}>{Math.round(maxRTT * n)}</span>
            ))}
          </div>
          <div className="plot-area">
            <svg
              viewBox="0 0 900 220"
              preserveAspectRatio="none"
              role="img"
              aria-label={`Latency chart showing ${points.length} probes. Latest: ${points.at(-1)?.success ? ms(points.at(-1)!.rtt) + " milliseconds" : "failed"}.`}
              onPointerLeave={() => setHover(null)}
              onPointerMove={(e) => {
                const rect = e.currentTarget.getBoundingClientRect();
                const target = ((e.clientX - rect.left) / rect.width) * 900;
                let nearest = 0;
                points.forEach((p, i) => {
                  if (
                    Math.abs(x(p) - target) <
                    Math.abs(x(points[nearest]) - target)
                  )
                    nearest = i;
                });
                setHover(nearest);
              }}
            >
              <defs>
                <linearGradient id="latency-area" x1="0" x2="0" y1="0" y2="1">
                  <stop
                    offset="0%"
                    stopColor="var(--accent)"
                    stopOpacity="0.34"
                  />
                  <stop
                    offset="65%"
                    stopColor="var(--accent)"
                    stopOpacity="0.08"
                  />
                  <stop
                    offset="100%"
                    stopColor="var(--accent)"
                    stopOpacity="0"
                  />
                </linearGradient>
                <filter
                  id="line-glow"
                  x="-10%"
                  y="-30%"
                  width="120%"
                  height="160%"
                >
                  <feGaussianBlur stdDeviation="3" result="blur" />
                  <feMerge>
                    <feMergeNode in="blur" />
                    <feMergeNode in="SourceGraphic" />
                  </feMerge>
                </filter>
              </defs>
              {[15, 63, 111, 159, 207].map((v) => (
                <line
                  key={v}
                  x1="0"
                  x2="900"
                  y1={v}
                  y2={v}
                  stroke="var(--border)"
                  strokeDasharray="2 7"
                  vectorEffect="non-scaling-stroke"
                />
              ))}
              {[8, 229, 450, 671, 892].map((v) => (
                <line
                  key={`vertical-${v}`}
                  x1={v}
                  x2={v}
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
              {points
                .filter((p) => !p.success || points.length < 3)
                .map((p) => (
                  <circle
                    key={p.sequence}
                    cx={x(p)}
                    cy={y(p)}
                    r="3.5"
                    fill={p.success ? "var(--accent)" : "var(--danger)"}
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
                style={{
                  left: `${Math.min(76, Math.max(3, x(hovered) / 9))}%`,
                }}
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
              <span>
                {clock(points[Math.floor((points.length - 1) / 2)].time)}
              </span>
              <span>{clock(points[points.length - 1].time)}</span>
            </div>
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
          {points.length} shown · {samples.length} retained
        </span>
      </div>
    </>
  );
}
