import { useCallback, useEffect, useRef, useState } from "react";
import { api, errorText, onUpdate, onRemove, type Detail, type Session } from "./api";

function append(detail: Detail, s: Session): Detail {
  const samples = [...detail.samples];
  if (s.last && samples.at(-1)?.sequence !== s.last.sequence) samples.push(s.last);
  return { session: s, samples: samples.slice(-6) };
}

export function useSessions(selected: string | null, search: string, filter: string, before: number) {
  const [sessions, setSessions] = useState<Session[]>([]);
  const [detail, setDetail] = useState<Detail | null>(null);
  const [error, setError] = useState("");
  const [connected, setConnected] = useState(false);
  const [loading, setLoading] = useState(false);
  const [next, setNext] = useState(0);
  const [active, setActive] = useState(0);
  const [refreshKey, setRefreshKey] = useState(0);
  const removed = useRef(new Set<string>());
  const refresh = useCallback(() => setRefreshKey(v => v + 1), []);
  const merge = useCallback((s: Session) => {
    if (removed.current.has(s.id)) return;
    setSessions(current => current.map(item => item.id === s.id && item.revision < s.revision ? s : item));
    setDetail(current => current?.session.id === s.id && current.session.revision < s.revision ? append(current, s) : current);
    if (s.saveError) setError(s.saveError);
  }, []);

  useEffect(() => {
    let mounted = true;
    const pending = new Map<string, Session>();
    const off = onUpdate(s => { pending.set(s.id, s); });
    setLoading(true);
    const timer = setTimeout(() => {
      api.list(search, filter, before).then(page => {
        if (!mounted) return;
        setSessions(page.sessions.filter(s => !removed.current.has(s.id)).map(s => {
          const recent = pending.get(s.id);
          return recent && recent.revision > s.revision ? recent : s;
        }));
        setNext(page.next); setActive(page.active); setConnected(true);
      }).catch(err => { if (mounted) setError(errorText(err)); })
        .finally(() => { off(); if (mounted) setLoading(false); });
    }, 150);
    return () => { mounted = false; clearTimeout(timer); off(); };
  }, [search, filter, before, refreshKey]);

  useEffect(() => {
    const offUpdate = onUpdate(s => {
      merge(s);
      if (s.revision === 1 || !s.running) refresh();
    });
    const offRemove = onRemove(id => {
      removed.current.add(id);
      setSessions(current => current.filter(s => s.id !== id));
      setDetail(current => current?.session.id === id ? null : current);
      refresh();
    });
    return () => { offUpdate(); offRemove(); };
  }, [merge, refresh]);

  useEffect(() => {
    let mounted = true;
    setDetail(null);
    if (!selected) return;
    let pending: Session[] = [];
    const off = onUpdate(s => { if (s.id === selected) pending = [...pending, s].slice(-6); });
    api.get(selected).then(value => {
      if (!mounted || removed.current.has(selected)) return;
      for (const s of pending) if (s.revision > value.session.revision) value = append(value, s);
      setDetail(current => current?.session.id === selected && current.session.revision > value.session.revision ? current : value);
      if (value.session.saveError) setError(value.session.saveError);
    }).catch(err => { if (mounted) setError(errorText(err)); }).finally(off);
    return () => { mounted = false; off(); };
  }, [selected]);
  return { sessions, detail, connected, error, setError, merge, refresh, next, active, loading };
}
