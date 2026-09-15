import type { Device } from "../types/network";

type DeviceInspectorProps = {
  device?: Device;
};

export function DeviceInspector({ device }: DeviceInspectorProps) {
  return (
    <aside className="inspector" aria-label="Device inspector">
      {device ? (
        <>
          <div className="panelHeader">
            <span>Selected Device</span>
            <strong>{device.hostname || device.ip}</strong>
          </div>
          <dl>
            <div>
              <dt>IP</dt>
              <dd>{device.ip}</dd>
            </div>
            <div>
              <dt>Vendor</dt>
              <dd>{device.vendor || "Unknown"}</dd>
            </div>
            <div>
              <dt>Latency</dt>
              <dd>{device.latencyMs} ms</dd>
            </div>
            <div>
              <dt>MAC</dt>
              <dd>{device.mac || "Unavailable"}</dd>
            </div>
            <div>
              <dt>Type</dt>
              <dd>{device.type}</dd>
            </div>
          </dl>
          {device.hints?.length ? (
            <>
              <h2>Identification Hints</h2>
              <ul>
                {device.hints.map((hint) => (
                  <li key={hint}>{hint}</li>
                ))}
              </ul>
            </>
          ) : null}
          <h2>Open Ports</h2>
          {device.ports?.length ? (
            <ul>
              {device.ports.map((port) => (
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
  );
}
