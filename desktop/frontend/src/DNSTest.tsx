import { useState, type FormEvent } from "react";
import { Play, Plus } from "lucide-react";
import type { Config } from "./api";

export function DNSForm({ disabled, busy, onStart }: {
  disabled: boolean; busy: boolean; onStart: (config: Config, resolvers: string) => Promise<void>;
}) {
  const [resolvers, setResolvers] = useState("");
  const [query, setQuery] = useState("");
  const [protocol, setProtocol] = useState<"udp" | "tcp">("udp");
  const [interval, setInterval] = useState(1);
  const [timeout, setTimeout] = useState(4);
  const [recording, setRecording] = useState(false);
  const submit = (event: FormEvent) => {
    event.preventDefault();
    void onStart({ type: "dns", url: "", method: "", query, protocol, recording,
      intervalMs: interval * 1000, timeoutMs: timeout * 1000, acceptedStatuses: [],
      proxy: { enabled: false, url: "", username: "", password: "" } }, resolvers);
  };
  return <form className="test-form panel" onSubmit={submit}>
    <div className="form-heading">
      <span className="section-icon"><Plus size={16} /></span>
      <h2>New test</h2>
      <label className="recording-option"><input type="checkbox" checked={recording} onChange={e => setRecording(e.target.checked)} />Result Recording {recording ? "ON" : "OFF"}</label>
    </div>
    <div className="dns-form-fields">
      <label className="dns-resolvers">DNS Server IP(s)
        <textarea required rows={3} maxLength={65536} placeholder={"One resolver IP per line\n1.1.1.1\n8.8.8.8"} value={resolvers} onChange={e => setResolvers(e.target.value)} spellCheck={false} />
      </label>
      <label className="dns-query">DNS Query
        <input required maxLength={4096} placeholder="Domain name to query" value={query} onChange={e => setQuery(e.target.value)} spellCheck={false} autoCapitalize="none" />
      </label>
      <label>DNS Protocol<select value={protocol} onChange={e => setProtocol(e.target.value as "udp" | "tcp")}><option value="udp">udp</option><option value="tcp">tcp</option></select></label>
      <label>Interval (s)<input required type="number" min={1} step={1} value={interval} onChange={e => setInterval(Number(e.target.value))} /></label>
      <label>Timeout (s)<input required type="number" min={1} step={1} value={timeout} onChange={e => setTimeout(Number(e.target.value))} /></label>
      <button className="button primary start-button" type="submit" disabled={disabled}><Play size={15} fill="currentColor" />{busy ? "Starting…" : "Start test"}</button>
    </div>
  </form>;
}
