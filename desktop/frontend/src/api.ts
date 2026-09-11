import { Call, Events } from "@wailsio/runtime";

export type Config = {
  type?: "http" | "dns" | "tcp" | "icmp";
  target?: string;
  payloadSize?: number;
  df?: boolean;
  port?: number;
  resolvedIP?: string;
  resolver?: string;
  query?: string;
  protocol?: "udp" | "tcp";
  recording?: boolean;
  url: string;
  method: string;
  intervalMs: number;
  timeoutMs: number;
  followRedirects?: boolean;
  acceptedStatuses: string[];
  proxy: {
    enabled: boolean;
    url: string;
    username: string;
    password: string;
  };
};
export type Sample = {
  dnsResponse?: string;
  dnsRecord?: string;
  sequence: number;
  time: string;
  rtt: number;
  statusCode: number;
  responsePhase?: string;
  success: boolean;
  error: string;
};
export type Session = {
  index?: number;
  id: string;
  config: Config;
  running: boolean;
  startedAt: string;
  endedAt: string | null;
  revision: number;
  endReason: string;
  saveError: string;
  importNote?: string;
  passwordRequired: boolean;
  sent: number;
  succeeded: number;
  minRtt: number;
  maxRtt: number;
  avgRtt: number;
  last: Sample | null;
};
export type Detail = { session: Session; samples: Sample[] };

export type ActiveByType = Record<"http" | "dns" | "tcp" | "icmp", number>;
export type Page = { sessions: Session[]; next: number; active: number; capacity: number; activeByType: ActiveByType };
export type Timeline = {
  samples: Sample[]; count: number; succeeded: number; average: number;
  maximum: number; from: number; to: number; revision: number; aggregated: boolean;
};

const call = <T>(method: string, ...args: unknown[]): Promise<T> =>
  Call.ByName(`main.TestService.${method}`, ...args);
export const api = {
  list: (search: string, filter: string, before: number) => call<Page>("List", search, filter, before),
  start: (config: Config) => call<Session>("Start", config),
  startDNS: (config: Config, resolvers: string) => call<Session[]>("StartDNS", config, resolvers),
  startICMP: (config: Config, targets: string) => call<Session[]>("StartICMP", config, targets),
  startTCP: (config: Config, targets: string) => call<Session[]>("StartTCP", config, targets),
  record: (id: string) => call<Session>("Record", id),
  dismiss: (id: string) => call<void>("Dismiss", id),
  restart: (id: string, password = "") => call<Session>("Restart", id, password),
  timeline: (id: string, from: number, to: number) => call<Timeline>("Timeline", id, from, to),
  exportCSV: (id: string) => call<string>("ExportCSV", id),
  importCSV: () => call<Session | null>("ImportCSV"),
  exportChart: (id: string, dataURL: string) => call<string>("ExportChart", id, dataURL),
  stop: (id: string) => call<Session>("Stop", id),
  get: (id: string) => call<Detail>("Get", id),
  remove: (id: string) => call<void>("Remove", id),
  chart: (id: string) => call<void>("OpenChart", id),
};
export const onUpdate = (callback: (s: Session) => void) =>
  Events.On("test:updated", (e) => callback(e.data as Session));
export const onRemove = (callback: (id: string) => void) =>
  Events.On("test:removed", (e) => callback(e.data as string));
export const errorText = (error: unknown) =>
  error instanceof Error ? error.message : String(error);
