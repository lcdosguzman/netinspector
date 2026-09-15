import {
  DeviceType,
  type ScanSummary as ProtoScanSummary,
  type ScanResult,
  ScanEventType,
  ScanMode as ProtoScanMode
} from "../src/gen/network/v1/network_pb";
import type { ScanSummary, ScanView } from "../types/network";

export function scanResultToView(result: ScanResult): ScanView {
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

export function scanSummaryToView(summary: ProtoScanSummary): ScanSummary {
  return {
    id: summary.id,
    mode: summary.mode === ProtoScanMode.REAL ? "REAL" : "DEMO",
    network: summary.network,
    startedAt: Number(summary.startedAt),
    endedAt: Number(summary.endedAt),
    deviceCount: summary.deviceCount
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
