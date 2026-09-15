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
import { useMemo, type ReactNode } from "react";
import type { Device } from "../types/network";

type TopologyNode = Node<{ label: ReactNode }>;

type TopologyGraphProps = {
  devices: Device[];
  selectedIP?: string;
  selectedDevice?: Device;
  onSelectDevice: (ip: string) => void;
};

export function TopologyGraph({ devices, selectedIP, selectedDevice, onSelectDevice }: TopologyGraphProps) {
  const topology = useMemo(() => buildTopology(devices, selectedIP), [devices, selectedIP]);

  const handleNodeClick: NodeMouseHandler<TopologyNode> = (_event, node) => {
    onSelectDevice(node.id);
  };

  return (
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
