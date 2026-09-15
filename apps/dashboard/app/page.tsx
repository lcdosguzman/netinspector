"use client";

import "@xyflow/react/dist/style.css";
import { useCallback, useEffect, useMemo, useState } from "react";
import { DeviceInspector } from "../components/DeviceInspector";
import { EventConsole } from "../components/EventConsole";
import { NetworkSummary } from "../components/NetworkSummary";
import { ScanControls } from "../components/ScanControls";
import { ScanHistory } from "../components/ScanHistory";
import { TopologyGraph } from "../components/TopologyGraph";
import { scanResultToView, scanSummaryToView } from "../lib/mappers";
import { getStoredScan, listScanHistory, startNetworkScan } from "../lib/netinspector-client";
import type { ScanMode, ScanStatus, ScanSummary, ScanView } from "../types/network";

const emptyScan: ScanView = {
  id: "pending",
  mode: "DEMO",
  network: "Not scanned yet",
  devices: [],
  events: []
};

export default function Home() {
  const [mode, setMode] = useState<ScanMode>("DEMO");
  const [scan, setScan] = useState<ScanView>(emptyScan);
  const [selectedIP, setSelectedIP] = useState<string>();
  const [status, setStatus] = useState<ScanStatus>("idle");
  const [error, setError] = useState<string>();
  const [history, setHistory] = useState<ScanSummary[]>([]);
  const [historyQuery, setHistoryQuery] = useState("");
  const [isHistoryLoading, setIsHistoryLoading] = useState(false);

  const selectedDevice = useMemo(() => {
    return scan.devices.find((device) => device.ip === selectedIP) ?? scan.devices[0];
  }, [scan.devices, selectedIP]);

  const refreshHistory = useCallback(async (query: string) => {
    setIsHistoryLoading(true);
    try {
      const response = await listScanHistory(query);
      setHistory(response.scans.map(scanSummaryToView));
    } catch {
      setHistory([]);
    } finally {
      setIsHistoryLoading(false);
    }
  }, []);

  useEffect(() => {
    void refreshHistory("");
  }, [refreshHistory]);

  async function startScan() {
    setStatus("running");
    setError(undefined);

    try {
      const result = await startNetworkScan(mode);
      const nextScan = scanResultToView(result);
      setScan(nextScan);
      setSelectedIP(nextScan.devices[0]?.ip);
      setStatus("success");
      if (nextScan.mode === "REAL") {
        await refreshHistory(historyQuery);
      }
    } catch (scanError) {
      setStatus("error");
      setError(scanError instanceof Error ? scanError.message : "Unexpected scan error");
    }
  }

  async function openStoredScan(id: string) {
    setStatus("running");
    setError(undefined);

    try {
      const result = await getStoredScan(id);
      const storedScan = scanResultToView(result);
      setScan(storedScan);
      setSelectedIP(storedScan.devices[0]?.ip);
      setStatus("success");
    } catch (scanError) {
      setStatus("error");
      setError(scanError instanceof Error ? scanError.message : "Unexpected scan history error");
    }
  }

  function updateHistoryQuery(query: string) {
    setHistoryQuery(query);
    void refreshHistory(query);
  }

  const statusLabel = {
    idle: `${mode === "DEMO" ? "Demo" : "Real"} Ready`,
    running: `Running ${mode.toLowerCase()} scan`,
    success: `${scan.mode === "DEMO" ? "Demo" : "Real"} Scan Complete`,
    error: "Scan Failed"
  }[status];

  return (
    <main className="shell">
      <header className="topbar">
        <div>
          <p className="eyebrow">NetInspector</p>
          <h1>Local Network Topology</h1>
        </div>
        <ScanControls mode={mode} status={status} onModeChange={setMode} onStartScan={startScan} />
      </header>

      <NetworkSummary deviceCount={scan.devices.length} network={scan.network} statusLabel={statusLabel} />

      {error ? <p className="banner">{error}</p> : null}

      <section className="workspace">
        <TopologyGraph
          devices={scan.devices}
          onSelectDevice={setSelectedIP}
          selectedDevice={selectedDevice}
          selectedIP={selectedIP}
        />
        <DeviceInspector device={selectedDevice} />
      </section>

      <ScanHistory
        activeScanID={scan.id}
        history={history}
        isLoading={isHistoryLoading}
        onOpenScan={openStoredScan}
        onQueryChange={updateHistoryQuery}
        onRefresh={() => refreshHistory(historyQuery)}
        query={historyQuery}
      />

      <EventConsole events={scan.events} />
    </main>
  );
}
