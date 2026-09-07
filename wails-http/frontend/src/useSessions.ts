import { useCallback, useEffect, useRef, useState } from "react";
import {
  api,
  errorText,
  onRemove,
  onUpdate,
  type Detail,
  type Session,
} from "./api";

export function useSessions(selected: string | null) {
  const [sessions, setSessions] = useState<Session[]>([]);
  const [detail, setDetail] = useState<Detail | null>(null);
  const [error, setError] = useState("");
  const [connected, setConnected] = useState(false);
  const removed = useRef(new Set<string>());
  const merge = useCallback((s: Session) => {
    if (removed.current.has(s.id)) return;
    setSessions((current) => {
      const found = current.find((item) => item.id === s.id);
      if (found && found.revision >= s.revision) return current;
      return found
        ? current.map((item) => (item.id === s.id ? s : item))
        : [...current, s];
    });
    setDetail((current) => {
      if (
        !current ||
        current.session.id !== s.id ||
        current.session.revision >= s.revision
      )
        return current;
      if (
        s.last &&
        current.samples.at(-1)?.sequence !== s.last.sequence
      )
        current.samples.push(s.last);
      return { session: s, samples: current.samples };
    });
  }, []);
  useEffect(() => {
    let mounted = true;
    const offUpdate = onUpdate(merge);
    const offRemove = onRemove((id) => {
      removed.current.add(id);
      setSessions((current) => current.filter((s) => s.id !== id));
      setDetail((current) => (current?.session.id === id ? null : current));
    });
    api
      .list()
      .then((items) => {
        if (mounted) {
          items.forEach(merge);
          setConnected(true);
        }
      })
      .catch((err) => {
        if (mounted)
          setError(
            `Desktop connection unavailable. Open the compiled Wails app. ${errorText(err)}`,
          );
      });
    return () => {
      mounted = false;
      offUpdate();
      offRemove();
    };
  }, [merge]);
  useEffect(() => {
    let mounted = true;
    setDetail(null);
    if (!selected) return;
    // Reconcile events arriving while the initial detail snapshot is in flight.
    const pending: Session[] = [];
    const off = onUpdate((s) => {
      if (s.id === selected) pending.push(s);
    });
    api
      .get(selected)
      .then((value) => {
        if (!mounted || removed.current.has(selected)) return;
        for (const s of pending) {
          if (s.revision <= value.session.revision) continue;
          if (s.last && value.samples.at(-1)?.sequence !== s.last.sequence)
            value.samples.push(s.last);
          value.session = s;
        }
        setDetail(value);
        merge(value.session);
      })
      .catch((err) => {
        if (mounted) setError(errorText(err));
      })
      .finally(off);
    return () => {
      mounted = false;
      off();
    };
  }, [selected, merge]);
  return { sessions, detail, connected, error, setError, merge };
}
