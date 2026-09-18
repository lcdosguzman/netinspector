import { createPromiseClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { NetworkService } from "../src/gen/network/v1/network_connect";
import {
  GetScanRequest,
  ListScansRequest,
  ScanMode as ProtoScanMode,
  StartScanRequest
} from "../src/gen/network/v1/network_pb";
import type { ScanMode } from "../types/network";

const apiBaseURL = process.env.NEXT_PUBLIC_LANSWEEPGO_API_URL ?? "http://127.0.0.1:8088";
const transport = createConnectTransport({ baseUrl: apiBaseURL });
const client = createPromiseClient(NetworkService, transport);

export function startNetworkScan(mode: ScanMode) {
  return client.startScan(new StartScanRequest({
    mode: mode === "REAL" ? ProtoScanMode.REAL : ProtoScanMode.DEMO
  }));
}

export function listScanHistory(query: string, limit = 20) {
  return client.listScans(new ListScansRequest({ query, limit }));
}

export function getStoredScan(id: string) {
  return client.getScan(new GetScanRequest({ id }));
}

export function scanExportURL(id: string, format: "json" | "csv") {
  return `${apiBaseURL}/api/scans/${encodeURIComponent(id)}/export/${format}`;
}
