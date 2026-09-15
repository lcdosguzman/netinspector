"use client";

import { createPromiseClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import {
  Background,
  Controls,
  MiniMap,
  ReactFlow,
  type Edge,
  MarkerType,
  type Node,
  type NodeMouseHandler
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";
import { useMemo, useState, type ReactNode } from "react";
import { NetworkService } from "../src/gen/network/v1/network_connect";
import {
  DeviceType,
  type ScanResult,
  ScanEventType,
  ScanMode as ProtoScanMode,
  StartScanRequest
} from "../src/gen/network/v1/network_pb";

type ScanMode = "DEMO" | "REAL";
type ScanStatus = "idle" | "running" | "success" | "error";

type Port = {
  number: number;
  protocol: string;
  serviceName: string;
};

type Device = {
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

type ScanEvent = {
  type: string;
  message: string;
  timestamp: number;
};

type ScanView = {
  id: string;
  mode: ScanMode;
  network: string;
  devices: Device[];
  events: ScanEvent[];
};

type TopologyNode = Node<{ label: ReactNode }>;

const apiBaseURL = process.env.NEXT_PUBLIC_NETINSPECTOR_API_URL ?? "http://127.0.0.1:8088";
const transport = createConnectTransport({ baseUrl: apiBaseURL });
const client = createPromiseClient(NetworkService, transport);

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

  const selectedDevice = useMemo(() => {
    return scan.devices.find((device) => device.ip === selectedIP) ?? scan.devices[0];
  }, [scan.devices, selectedIP]);

  const topology = useMemo(() => buildTopology(scan.devices, selectedIP), [scan.devices, selectedIP]);

  const handleNodeClick: NodeMouseHandler<TopologyNode> = (_event, node) => {
    setSelectedIP(node.id);
  };

  async function startScan() {
    setStatus("running");
    setError(undefined);

    try {
      const result = await client.startScan(new StartScanRequest({
        mode: mode === "REAL" ? ProtoScanMode.REAL : ProtoScanMode.DEMO
      }));
      const nextScan = scanResultToView(result);
      setScan(nextScan);
      setSelectedIP(nextScan.devices[0]?.ip);
      setStatus("success");
    } catch (scanError) {
      setStatus("error");
      setError(scanError instanceof Error ? scanError.message : "Unexpected scan error");
    }
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
        <div className="actions" aria-label="Scan controls">
          <button
            className={`segmented ${mode === "DEMO" ? "active" : ""}`}
            disabled={status === "running"}
            onClick={() => setMode("DEMO")}
            type="button"
          >
            Demo
          </button>
          <button
            className={`segmented ${mode === "REAL" ? "active" : ""}`}
            disabled={status === "running"}
            onClick={() => setMode("REAL")}
            type="button"
          >
            Real
          </button>
          <button className="primary" disabled={status === "running"} onClick={startScan} type="button">
            {status === "running" ? "Scanning..." : "Start Scan"}
          </button>
        </div>
      </header>

      <section className="summary" aria-label="Network summary">
        <div>
          <span>Network</span>
          <strong>{scan.network}</strong>
        </div>
        <div>
          <span>Devices</span>
          <strong>{scan.devices.length}</strong>
        </div>
        <div>
          <span>Status</span>
          <strong>{statusLabel}</strong>
        </div>
      </section>

      {error ? <p className="banner">{error}</p> : null}

      <section className="workspace">
        <div className="topology" aria-label="Interactive topology graph">
          {selectedDevice ? (
            <ReactFlow<TopologyNode, Edge>
              className="topologyGraph"
              colorMode="dark"
              edges={topology.edges}
              fitView
              fitViewOptions={{ padding: 0.25 }}
              maxZoom={1.6}
              minZoom={0.45}
              nodes={topology.nodes}
              nodesConnectable={false}
              onNodeClick={handleNodeClick}
              panOnScroll
              proOptions={{ hideAttribution: true }}
            >
              <Background color="rgba(154, 166, 178, 0.22)" gap={28} />
              <MiniMap
                maskColor="rgba(17, 19, 24, 0.68)"
                nodeColor={(node) => (node.id === selectedDevice.ip ? "#72e0bf" : "#303846")}
                nodeStrokeWidth={3}
                pannable
                zoomable
              />
              <Controls showInteractive={false} />
            </ReactFlow>
          ) : (
            <div className="emptyState">
              <strong>No scan data yet</strong>
              <span>Select Demo or Real, then start a scan.</span>
            </div>
          )}
        </div>

        <aside className="inspector" aria-label="Device inspector">
          {selectedDevice ? (
            <>
              <div className="panelHeader">
                <span>Selected Device</span>
                <strong>{selectedDevice.hostname || selectedDevice.ip}</strong>
              </div>
              <dl>
                <div>
                  <dt>IP</dt>
                  <dd>{selectedDevice.ip}</dd>
                </div>
                <div>
                  <dt>Vendor</dt>
                  <dd>{selectedDevice.vendor || "Unknown"}</dd>
                </div>
                <div>
                  <dt>Latency</dt>
                  <dd>{selectedDevice.latencyMs} ms</dd>
                </div>
                <div>
                  <dt>MAC</dt>
                  <dd>{selectedDevice.mac || "Unavailable"}</dd>
                </div>
                <div>
                  <dt>Type</dt>
                  <dd>{selectedDevice.type}</dd>
                </div>
              </dl>
              {selectedDevice.hints?.length ? (
                <>
                  <h2>Identification Hints</h2>
                  <ul>
                    {selectedDevice.hints.map((hint) => (
                      <li key={hint}>{hint}</li>
                    ))}
                  </ul>
                </>
              ) : null}
              <h2>Open Ports</h2>
              {selectedDevice.ports?.length ? (
                <ul>
                  {selectedDevice.ports.map((port) => (
                    <li key={`${port.protocol}-${port.number}`}>
                      {port.number} {port.serviceName}
                    </li>
                  ))}
                </ul>
              ) : (
                <p className="muted">No open ports reported.</p>
              )}
            </>
          ) : (
            <p className="muted">No device selected.</p>
          )}
        </aside>
      </section>

      <section className="console" aria-label="Scan event console">
        {scan.events.length ? (
          scan.events.map((event) => (
            <p key={`${event.timestamp}-${event.message}`}>
              [{event.type}] {event.message}
            </p>
          ))
        ) : (
          <p>[system] Waiting for scan.</p>
        )}
      </section>
    </main>
  );
}

function buildTopology(devices: Device[], selectedIP?: string): { nodes: TopologyNode[]; edges: Edge[] } {
  if (!devices.length) {
    return { nodes: [], edges: [] };
  }

  const gateway = devices[0];
  const leafDevices = devices.slice(1);
  const radius = Math.max(190, Math.min(330, 150 + leafDevices.length * 26));
  const nodes: TopologyNode[] = [
    {
      id: gateway.ip,
      type: "default",
      position: { x: 0, y: 0 },
      className: `flowNode gatewayNode ${gateway.ip === selectedIP ? "selectedNode" : ""}`,
      data: { label: <TopologyNodeLabel device={gateway} /> }
    }
  ];

  leafDevices.forEach((device, index) => {
    const angle = (index / Math.max(leafDevices.length, 1)) * Math.PI * 2 - Math.PI / 2;
    nodes.push({
      id: device.ip,
      type: "default",
      position: {
        x: Math.cos(angle) * radius,
        y: Math.sin(angle) * radius
      },
      className: `flowNode deviceNode ${device.ip === selectedIP ? "selectedNode" : ""}`,
      data: { label: <TopologyNodeLabel device={device} /> }
    });
  });

  const edges: Edge[] = leafDevices.map((device) => ({
    id: `${gateway.ip}-${device.ip}`,
    source: gateway.ip,
    target: device.ip,
    animated: device.ip === selectedIP,
    className: device.ip === selectedIP ? "selectedEdge" : undefined,
    markerEnd: {
      type: MarkerType.ArrowClosed,
      color: device.ip === selectedIP ? "#72e0bf" : "#526070"
    },
    style: {
      stroke: device.ip === selectedIP ? "#72e0bf" : "#526070",
      strokeWidth: device.ip === selectedIP ? 2.5 : 1.5
    }
  }));

  return { nodes, edges };
}

function TopologyNodeLabel({ device }: { device: Device }) {
  return (
    <div className="flowNodeLabel">
      <span>{device.type}</span>
      <strong>{device.hostname || device.ip}</strong>
      {device.hostname ? <small>{device.ip}</small> : null}
      <small>{device.vendor || "Unknown vendor"}</small>
    </div>
  );
}

function scanResultToView(result: ScanResult): ScanView {
  return {
    id: result.id,
    mode: result.mode === ProtoScanMode.REAL ? "REAL" : "DEMO",
    network: result.network,
    devices: result.devices.map((device) => ({
      ip: device.ip,
      hostname: device.hostname,
      mac: device.mac,
      vendor: device.vendor,
      type: deviceTypeLabel(device.deviceType),
      hints: device.hints,
      latencyMs: Number(device.latencyMs),
      isActive: device.isActive,
      ports: device.ports.map((port) => ({
        number: port.port,
        protocol: port.protocol,
        serviceName: port.serviceName
      }))
    })),
    events: result.events.map((event) => ({
      type: scanEventTypeLabel(event.eventType),
      message: event.message,
      timestamp: Number(event.timestamp)
    }))
  };
}

function deviceTypeLabel(deviceType: DeviceType): string {
  return {
    [DeviceType.UNSPECIFIED]: "UNKNOWN",
    [DeviceType.ROUTER]: "ROUTER",
    [DeviceType.DESKTOP]: "DESKTOP",
    [DeviceType.MOBILE]: "MOBILE",
    [DeviceType.TV]: "TV",
    [DeviceType.PRINTER]: "PRINTER",
    [DeviceType.IOT]: "IOT",
    [DeviceType.UNKNOWN]: "UNKNOWN"
  }[deviceType];
}

function scanEventTypeLabel(eventType: ScanEventType): string {
  return {
    [ScanEventType.UNSPECIFIED]: "event",
    [ScanEventType.SCAN_STARTED]: "scan_started",
    [ScanEventType.DEVICE_DISCOVERED]: "device_discovered",
    [ScanEventType.DEVICE_UPDATED]: "device_updated",
    [ScanEventType.SCAN_FINISHED]: "scan_finished",
    [ScanEventType.SCAN_FAILED]: "scan_failed"
  }[eventType];
}
