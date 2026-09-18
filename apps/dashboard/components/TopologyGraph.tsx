import {
  Background,
  Controls,
  MiniMap,
  ReactFlow,
  useEdgesState,
  useNodesState,
  type Edge,
  MarkerType,
  type Node,
  type NodeMouseHandler
} from "@xyflow/react";
import { useEffect, useMemo, type ReactNode } from "react";
import type { Device } from "../types/network";

type TopologyNode = Node<{ label: ReactNode; deviceType: string; isGateway: boolean }>;

type TopologyGraphProps = {
  devices: Device[];
  selectedIP?: string;
  selectedDevice?: Device;
  onSelectDevice: (ip: string) => void;
};

export function TopologyGraph({ devices, selectedIP, selectedDevice, onSelectDevice }: TopologyGraphProps) {
  const topology = useMemo(() => buildTopology(devices, selectedIP), [devices, selectedIP]);
  const [nodes, setNodes, onNodesChange] = useNodesState<TopologyNode>(topology.nodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>(topology.edges);

  useEffect(() => {
    setNodes((currentNodes) => mergeNodePositions(topology.nodes, currentNodes));
    setEdges(topology.edges);
  }, [setEdges, setNodes, topology]);

  const handleNodeClick: NodeMouseHandler<TopologyNode> = (_event, node) => {
    onSelectDevice(node.id);
  };

  return (
    <div className="topology" aria-label="Interactive topology graph">
      {selectedDevice ? (
        <ReactFlow<TopologyNode, Edge>
          className="topologyGraph"
          colorMode="dark"
          edges={edges}
          fitView
          fitViewOptions={{ padding: 0.25 }}
          maxZoom={1.6}
          minZoom={0.45}
          nodes={nodes}
          nodesConnectable={false}
          nodesDraggable
          onEdgesChange={onEdgesChange}
          onNodesChange={onNodesChange}
          onNodeClick={handleNodeClick}
          panOnDrag={[1, 2]}
          panOnScroll
          proOptions={{ hideAttribution: true }}
        >
          <Background color="rgba(154, 166, 178, 0.22)" gap={28} />
          <MiniMap
            maskColor="rgba(17, 19, 24, 0.68)"
            nodeColor={(node) => (node.id === selectedDevice.ip ? "#72e0bf" : minimapNodeColor(node as TopologyNode))}
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

  const gateway = devices.find((device) => device.type === "ROUTER") ?? devices[0];
  const leafDevices = devices
    .filter((device) => device.ip !== gateway.ip)
    .sort((left, right) => deviceSortLabel(left).localeCompare(deviceSortLabel(right)));
  const nodes: TopologyNode[] = [
    {
      id: gateway.ip,
      type: "default",
      position: { x: 0, y: 0 },
      className: `flowNode gatewayNode type-${deviceTypeClass(gateway.type)} ${gateway.ip === selectedIP ? "selectedNode" : ""}`,
      data: { label: <TopologyNodeLabel device={gateway} isGateway />, deviceType: gateway.type, isGateway: true }
    }
  ];

  const groups = groupDevicesByType(leafDevices);
  const columnGap = 250;
  const rowGap = 150;
  const startX = 330;
  const maxRows = Math.max(...groups.map((group) => group.devices.length), 1);
  const gatewayOffsetY = ((maxRows - 1) * rowGap) / 2;
  nodes[0].position = { x: 0, y: gatewayOffsetY };

  groups.forEach((group, columnIndex) => {
    const columnX = startX + columnIndex * columnGap;
    const columnHeight = (group.devices.length - 1) * rowGap;
    const startY = gatewayOffsetY - columnHeight / 2;

    group.devices.forEach((device, rowIndex) => {
      nodes.push({
        id: device.ip,
        type: "default",
        position: {
          x: columnX,
          y: startY + rowIndex * rowGap
        },
        className: `flowNode deviceNode type-${deviceTypeClass(device.type)} ${device.ip === selectedIP ? "selectedNode" : ""}`,
        data: { label: <TopologyNodeLabel device={device} />, deviceType: device.type, isGateway: false }
      });
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

function mergeNodePositions(nextNodes: TopologyNode[], currentNodes: TopologyNode[]) {
  const currentByID = new Map(currentNodes.map((node) => [node.id, node]));

  return nextNodes.map((nextNode) => {
    const currentNode = currentByID.get(nextNode.id);
    if (!currentNode) {
      return nextNode;
    }

    return {
      ...nextNode,
      position: currentNode.position
    };
  });
}

function groupDevicesByType(devices: Device[]) {
  const typeOrder = ["DESKTOP", "MOBILE", "TV", "PRINTER", "IOT", "ROUTER", "UNKNOWN"];
  const grouped = new Map<string, Device[]>();

  for (const device of devices) {
    const type = typeOrder.includes(device.type) ? device.type : "UNKNOWN";
    grouped.set(type, [...(grouped.get(type) ?? []), device]);
  }

  return typeOrder
    .filter((type) => grouped.has(type))
    .map((type) => ({
      type,
      devices: (grouped.get(type) ?? []).sort((left, right) => deviceSortLabel(left).localeCompare(deviceSortLabel(right)))
    }));
}

function deviceSortLabel(device: Device) {
  return `${device.type}-${device.hostname || device.vendor || device.ip}`;
}

function deviceTypeClass(type: string) {
  return type.toLowerCase().replace(/[^a-z0-9]+/g, "-") || "unknown";
}

function minimapNodeColor(node: TopologyNode) {
  if (node.data.isGateway) {
    return "#51c7a8";
  }
  return {
    DESKTOP: "#7aa2f7",
    MOBILE: "#f2bb5f",
    TV: "#b48ead",
    PRINTER: "#8bd5ca",
    IOT: "#a3be8c",
    ROUTER: "#51c7a8",
    UNKNOWN: "#526070"
  }[node.data.deviceType] ?? "#526070";
}

function TopologyNodeLabel({ device, isGateway = false }: { device: Device; isGateway?: boolean }) {
  const portCount = device.ports?.length ?? 0;

  return (
    <div className="flowNodeLabel">
      <div className="nodeMeta">
        <span>{isGateway ? "GATEWAY" : device.type}</span>
        <i aria-hidden="true" />
      </div>
      <strong>{device.hostname || device.ip}</strong>
      {device.hostname ? <small>{device.ip}</small> : null}
      <small>{device.vendor || "Unknown vendor"}</small>
      <small>{portCount ? `${portCount} open ${portCount === 1 ? "port" : "ports"}` : "No open ports found"}</small>
    </div>
  );
}
