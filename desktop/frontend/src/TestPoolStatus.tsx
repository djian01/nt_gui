import type { ActiveByType } from "./api";

const types = ["http", "dns", "tcp", "icmp"] as const;

export function TestPoolStatus({ active, capacity, activeByType }: {
  active: number;
  capacity: number;
  activeByType: ActiveByType;
}) {
  const occupied = types.flatMap(type => Array.from({ length: activeByType[type] }, () => type));
  const empty = Math.max(0, capacity - active);
  return (
    <section className="test-pool" aria-label="Shared concurrent test slots" role="status">
      <strong className="test-pool-total">{capacity ? `${active}/${capacity} Tests running` : "Loading test slots…"}</strong>
      <div className="test-pool-slots" aria-hidden="true">
        {Array.from({ length: capacity }, (_, index) => (
          <i key={index} className={occupied[index] ? `pool-${occupied[index]}` : "pool-empty"}
            title={occupied[index] ? `${occupied[index].toUpperCase()} running` : "Empty slot"} />
        ))}
      </div>
      <div className="test-pool-legend">
        {types.map(type => <span key={type}><i className={`pool-${type}`} aria-hidden="true" />{type.toUpperCase()} <b>{activeByType[type]}</b></span>)}
      </div>
      <span className="test-pool-empty">{capacity ? `${empty} empty ${empty === 1 ? "slot" : "slots"}` : ""}</span>
    </section>
  );
}
