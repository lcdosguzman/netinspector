import type { ScanSummary } from "../types/network";
import { scanExportURL } from "../lib/lansweepgo-client";

type ScanHistoryProps = {
  history: ScanSummary[];
  query: string;
  activeScanID: string;
  isLoading: boolean;
  onQueryChange: (query: string) => void;
  onOpenScan: (id: string) => void;
  onRefresh: () => void;
};

export function ScanHistory({
  history,
  query,
  activeScanID,
  isLoading,
  onQueryChange,
  onOpenScan,
  onRefresh
}: ScanHistoryProps) {
  return (
    <section className="history" aria-label="Scan history">
      <div className="historyHeader">
        <div>
          <span>Scan History</span>
          <strong>{history.length ? `${history.length} saved scans` : "No saved scans"}</strong>
        </div>
        <button className="secondary" disabled={isLoading} onClick={onRefresh} type="button">
          Refresh
        </button>
      </div>
      <input
        aria-label="Search scan history"
        className="historySearch"
        onChange={(event) => onQueryChange(event.target.value)}
        placeholder="Search by IP, hostname, MAC, vendor, network..."
        type="search"
        value={query}
      />
      {history.length ? (
        <div className="historyList">
          {history.map((scan) => (
            <article className={`historyItem ${scan.id === activeScanID ? "active" : ""}`} key={scan.id}>
              <button className="historyOpen" onClick={() => onOpenScan(scan.id)} type="button">
                <span>{formatScanDate(scan.startedAt)}</span>
                <strong>{scan.network}</strong>
                <small>
                  {scan.mode} · {scan.deviceCount} devices
                </small>
              </button>
              <div className="historyExports" aria-label={`Export scan ${scan.id}`}>
                <a href={scanExportURL(scan.id, "json")}>JSON</a>
                <a href={scanExportURL(scan.id, "csv")}>CSV</a>
              </div>
            </article>
          ))}
        </div>
      ) : (
        <p className="muted">{isLoading ? "Loading scan history..." : "Run a real scan to save history."}</p>
      )}
    </section>
  );
}

function formatScanDate(timestamp: number): string {
  if (!timestamp) {
    return "Unknown date";
  }
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short"
  }).format(new Date(timestamp));
}
