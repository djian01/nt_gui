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
  let path = "";
  let previousSuccess = false;
  for (const point of points) {
    if (point.success)
      path += `${previousSuccess ? " L" : " M"}${x(point)},${y(point)}`;
    previousSuccess = point.success;
  }
  const hovered =
    hover == null ? null : points[Math.min(hover, points.length - 1)];
  return (
    <>
      <div className="chart-toolbar">
        <span>
          <i className="legend-line" /> Response time <small>ms</small>
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
              {[15, 63, 111, 159, 207].map((v) => (
                <line
                  key={v}
                  x1="0"
                  x2="900"
                  y1={v}
                  y2={v}
                  stroke="#e7edf0"
                  strokeDasharray="4 5"
                  vectorEffect="non-scaling-stroke"
                />
              ))}
              <path
                d={path}
                fill="none"
                stroke="#0a9684"
                strokeWidth="2.5"
                strokeLinejoin="round"
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
                    fill={p.success ? "#0a9684" : "#dc645a"}
                  />
                ))}
              {hovered && (
                <>
                  <line
                    x1={x(hovered)}
                    x2={x(hovered)}
                    y1="0"
                    y2="220"
                    stroke="#9caeb6"
                    strokeDasharray="4 4"
                  />
                  <circle
                    cx={x(hovered)}
                    cy={y(hovered)}
                    r="5"
                    fill={hovered.success ? "#0a9684" : "#dc645a"}
                    stroke="white"
                    strokeWidth="2"
                  />
                </>
              )}
            </svg>
            {hovered && (
              <div
                className="chart-tooltip"
                style={{
                  left: `${Math.min(76, Math.max(3, x(hovered) / 9))}%`,
                }}
              >
                <span>
                  #{hovered.sequence} · {clock(hovered.time)}
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
          <i className="dot teal" />
          Successful response <i className="dot red" />
          Failed probe
        </span>
        <span>
          {points.length} of {samples.length} retained probes
        </span>
      </div>
    </>
  );
}
