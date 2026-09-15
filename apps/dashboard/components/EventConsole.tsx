import type { ScanEvent } from "../types/network";

type EventConsoleProps = {
  events: ScanEvent[];
};

export function EventConsole({ events }: EventConsoleProps) {
  return (
    <section className="console" aria-label="Scan event console">
      {events.length ? (
        events.map((event) => (
          <p key={`${event.timestamp}-${event.message}`}>
            [{event.type}] {event.message}
          </p>
        ))
      ) : (
        <p>[system] Waiting for scan.</p>
      )}
    </section>
  );
}
