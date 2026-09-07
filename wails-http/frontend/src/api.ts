import { Call, Events } from "@wailsio/runtime";

export type Config = {
  url: string;
  method: string;
  intervalMs: number;
  timeoutMs: number;
  acceptedStatuses: string[];
  proxy: {
    enabled: boolean;
    url: string;
    username: string;
    password: string;
  };
};
export type Sample = {
  sequence: number;
  time: string;
  rtt: number;
  statusCode: number;
  success: boolean;
  error: string;
};
export type Session = {
  id: string;
  config: Config;
  running: boolean;
  startedAt: string;
  endedAt: string | null;
  revision: number;
  endReason: string;
  saveError: string;
  passwordRequired: boolean;
  sent: number;
  succeeded: number;
  minRtt: number;
  maxRtt: number;
  avgRtt: number;
  last: Sample | null;
};
export type Detail = { session: Session; samples: Sample[] };

export type Page = { sessions: Session[]; next: number; active: number };
export type Timeline = {
  samples: Sample[]; count: number; succeeded: number; average: number;
  maximum: number; from: number; to: number; revision: number; aggregated: boolean;
};

const call = <T>(method: string, ...args: unknown[]): Promise<T> =>
  Call.ByName(`main.PingService.${method}`, ...args);
export const api = {
  list: (search: string, filter: string, before: number) => call<Page>("List", search, filter, before),
  start: (config: Config) => call<Session>("Start", config),
  restart: (id: string, password = "") => call<Session>("Restart", id, password),
  timeline: (id: string, from: number, to: number) => call<Timeline>("Timeline", id, from, to),
  exportCSV: (id: string) => call<string>("ExportCSV", id),
  stop: (id: string) => call<Session>("Stop", id),
  get: (id: string) => call<Detail>("Get", id),
  remove: (id: string) => call<void>("Remove", id),
  chart: (id: string) => call<void>("OpenChart", id),
};
export const onUpdate = (callback: (s: Session) => void) =>
  Events.On("http:updated", (e) => callback(e.data as Session));
export const onRemove = (callback: (id: string) => void) =>
  Events.On("http:removed", (e) => callback(e.data as string));
export const errorText = (error: unknown) =>
  error instanceof Error ? error.message : String(error);
