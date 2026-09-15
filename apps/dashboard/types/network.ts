export type ScanMode = "DEMO" | "REAL";
export type ScanStatus = "idle" | "running" | "success" | "error";

export type Port = {
  number: number;
  protocol: string;
  serviceName: string;
};

export type Device = {
  ip: string;
  hostname?: string;
  mac?: string;
  vendor?: string;
  type: string;
  hints?: string[];
  latencyMs: number;
  isActive: boolean;
  ports?: Port[];
};

export type ScanEvent = {
  type: string;
  message: string;
  timestamp: number;
};

export type ScanView = {
  id: string;
  mode: ScanMode;
  network: string;
  devices: Device[];
  events: ScanEvent[];
};

export type ScanSummary = {
  id: string;
  mode: ScanMode;
  network: string;
  startedAt: number;
  endedAt: number;
  deviceCount: number;
};
