import { useState, type FormEvent } from "react";
import { Play, Plus } from "lucide-react";
import type { Config } from "./api";

export function TCPForm({ disabled, busy, onStart }: {
  disabled: boolean; busy: boolean; onStart: (config: Config, targets: string) => Promise<void>;
}) {
  const [targets, setTargets] = useState("");
  const [port, setPort] = useState("");
  const [interval, setInterval] = useState(1);
  const [timeout, setTimeout] = useState(4);
  const [recording, setRecording] = useState(false);
  const submit = (event: FormEvent) => {
    event.preventDefault();
    void onStart({ type: "tcp", url: "", method: "", port: Number(port), recording,
      intervalMs: interval * 1000, timeoutMs: timeout * 1000, acceptedStatuses: [],
      proxy: { enabled: false, url: "", username: "", password: "" } }, targets);
  };
  return <form className="test-form panel" onSubmit={submit}>
    <div className="form-heading">
      <span className="section-icon"><Plus size={16} /></span>
      <h2>New test</h2>
      <label className="recording-option"><input type="checkbox" checked={recording} onChange={e => setRecording(e.target.checked)} />Result Recording {recording ? "ON" : "OFF"}</label>
    </div>
    <div className="dns-form-fields tcp-form-fields">
      <label className="dns-resolvers">TCP Server IP/Host(s)
        <textarea required rows={3} maxLength={65536} placeholder={"One IP or hostname per line\nserver.example.com\n192.0.2.1"} value={targets} onChange={e => setTargets(e.target.value)} spellCheck={false} />
      </label>
      <label>TCP Port<input required type="number" min={1} max={65535} step={1} placeholder="1–65535" value={port} onChange={e => setPort(e.target.value)} /></label>
      <label>Interval (s)<input required type="number" min={1} step={1} value={interval} onChange={e => setInterval(Number(e.target.value))} /></label>
      <label>Timeout (s)<input required type="number" min={1} step={1} value={timeout} onChange={e => setTimeout(Number(e.target.value))} /></label>
      <button className="button primary start-button" type="submit" disabled={disabled}><Play size={15} fill="currentColor" />{busy ? "Starting…" : "Start test"}</button>
    </div>
  </form>;
}
