type NetworkSummaryProps = {
  network: string;
  deviceCount: number;
  statusLabel: string;
};

export function NetworkSummary({ network, deviceCount, statusLabel }: NetworkSummaryProps) {
  return (
    <section className="summary" aria-label="Network summary">
      <div>
        <span>Network</span>
        <strong>{network}</strong>
      </div>
      <div>
        <span>Devices</span>
        <strong>{deviceCount}</strong>
      </div>
      <div>
        <span>Status</span>
        <strong>{statusLabel}</strong>
      </div>
    </section>
  );
}
