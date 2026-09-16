import { describe, expect, it } from "vitest";
import { scanResultToView, scanSummaryToView } from "./mappers";
import {
  Device,
  DeviceType,
  PortInfo,
  ScanEvent,
  ScanEventType,
  ScanMode,
  ScanResult,
  ScanSummary
} from "../src/gen/network/v1/network_pb";

describe("scanResultToView", () => {
  const cases = [
    {
      name: "maps a real router scan",
      result: new ScanResult({
        id: "scan-real",
        mode: ScanMode.REAL,
        network: "192.168.1.0/24",
        devices: [
          new Device({
            ip: "192.168.1.1",
            hostname: "router.local",
            mac: "84:15:d3:1b:8e:dd",
            vendor: "Network equipment vendor",
            deviceType: DeviceType.ROUTER,
            latencyMs: BigInt(12),
            isActive: true,
            hints: ["DNS y HTTP abiertos; posible router o gateway."],
            ports: [
              new PortInfo({
                port: 80,
                protocol: "tcp",
                serviceName: "http"
              })
            ]
          })
        ],
        events: [
          new ScanEvent({
            eventType: ScanEventType.DEVICE_DISCOVERED,
            message: "Discovered 192.168.1.1",
            timestamp: BigInt(1789514933000)
          })
        ]
      }),
      wantMode: "REAL",
      wantDeviceType: "ROUTER",
      wantEventType: "device_discovered"
    },
    {
      name: "maps unspecified values to demo and unknown labels",
      result: new ScanResult({
        id: "scan-demo",
        mode: ScanMode.UNSPECIFIED,
        network: "demo",
        devices: [
          new Device({
            ip: "192.168.1.40",
            deviceType: DeviceType.UNSPECIFIED
          })
        ],
        events: [
          new ScanEvent({
            eventType: ScanEventType.UNSPECIFIED,
            message: "event"
          })
        ]
      }),
      wantMode: "DEMO",
      wantDeviceType: "UNKNOWN",
      wantEventType: "event"
    }
  ] as const;

  for (const test of cases) {
    it(test.name, () => {
      const view = scanResultToView(test.result);

      expect(view.mode).toBe(test.wantMode);
      expect(view.devices[0].type).toBe(test.wantDeviceType);
      expect(view.events[0].type).toBe(test.wantEventType);
    });
  }
});

describe("scanSummaryToView", () => {
  const cases = [
    {
      name: "maps real scan summary",
      summary: new ScanSummary({
        id: "scan-real",
        mode: ScanMode.REAL,
        network: "192.168.1.0/24",
        startedAt: BigInt(1789514933000),
        endedAt: BigInt(1789514940000),
        deviceCount: 7
      }),
      wantMode: "REAL"
    },
    {
      name: "maps unspecified summary mode to demo",
      summary: new ScanSummary({
        id: "scan-demo",
        mode: ScanMode.UNSPECIFIED
      }),
      wantMode: "DEMO"
    }
  ] as const;

  for (const test of cases) {
    it(test.name, () => {
      const view = scanSummaryToView(test.summary);

      expect(view.id).toBe(test.summary.id);
      expect(view.mode).toBe(test.wantMode);
      expect(view.deviceCount).toBe(test.summary.deviceCount);
    });
  }
});
