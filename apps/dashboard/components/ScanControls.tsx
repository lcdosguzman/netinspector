import type { ScanMode, ScanStatus } from "../types/network";

type ScanControlsProps = {
  mode: ScanMode;
  status: ScanStatus;
  onModeChange: (mode: ScanMode) => void;
  onStartScan: () => void;
};

export function ScanControls({ mode, status, onModeChange, onStartScan }: ScanControlsProps) {
  return (
    <div className="actions" aria-label="Scan controls">
      <button
        className={`segmented ${mode === "DEMO" ? "active" : ""}`}
        disabled={status === "running"}
        onClick={() => onModeChange("DEMO")}
        type="button"
      >
        Demo
      </button>
      <button
        className={`segmented ${mode === "REAL" ? "active" : ""}`}
        disabled={status === "running"}
        onClick={() => onModeChange("REAL")}
        type="button"
      >
        Real
      </button>
      <button className="primary" disabled={status === "running"} onClick={onStartScan} type="button">
        {status === "running" ? "Scanning..." : "Start Scan"}
      </button>
    </div>
  );
}
