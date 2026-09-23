export const VM_STATUS_STYLES: Record<string, string> = {
  running: "bg-bc-success/10 text-bc-success border-bc-success/30",
  provisioning: "bg-bc-warning/10 text-bc-warning border-bc-warning/30",
  pending: "bg-bc-warning/10 text-bc-warning border-bc-warning/30",
  starting: "bg-bc-warning/10 text-bc-warning border-bc-warning/30",
  stopping: "bg-bc-warning/10 text-bc-warning border-bc-warning/30",
  deleting: "bg-bc-warning/10 text-bc-warning border-bc-warning/30",
  stopped: "bg-bc-text-tertiary/10 text-bc-text-secondary border-bc-border",
  deleted: "bg-bc-text-tertiary/10 text-bc-text-tertiary border-bc-border",
  error: "bg-bc-danger/10 text-bc-danger border-bc-danger/30",
  online: "bg-bc-success/10 text-bc-success border-bc-success/30",
  offline: "bg-bc-text-tertiary/10 text-bc-text-tertiary border-bc-border",
};

export function statusBadgeClass(status: string): string {
  return (
    VM_STATUS_STYLES[status] ||
    "bg-bc-text-tertiary/10 text-bc-text-secondary border-bc-border"
  );
}

export function bytesToGiB(bytes: number): number {
  return Math.round((bytes / 1024 ** 3) * 10) / 10;
}

export function formatRelative(iso?: string): string {
  if (!iso) return "—";
  const then = new Date(iso).getTime();
  if (Number.isNaN(then)) return "—";
  const diff = Date.now() - then;
  const sec = Math.round(diff / 1000);
  if (sec < 60) return `hace ${sec}s`;
  const min = Math.round(sec / 60);
  if (min < 60) return `hace ${min}m`;
  const hrs = Math.round(min / 60);
  if (hrs < 24) return `hace ${hrs}h`;
  const days = Math.round(hrs / 24);
  return `hace ${days}d`;
}
