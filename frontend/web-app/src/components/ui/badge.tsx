export function Badge({
  children,
  variant = "default",
}: {
  children: React.ReactNode;
  variant?: "default" | "primary" | "success" | "warning" | "danger";
}) {
  const styles: Record<string, string> = {
    default: "border-bc-border bg-bc-card text-bc-text-secondary",
    primary: "border-bc-primary/40 bg-bc-primary/10 text-bc-primary",
    success: "border-bc-success/30 bg-bc-success/10 text-bc-success",
    warning: "border-bc-warning/30 bg-bc-warning/10 text-bc-warning",
    danger: "border-bc-danger/30 bg-bc-danger/10 text-bc-danger",
  };
  return (
    <span
      className={`inline-flex items-center rounded-full border px-3 py-1 text-[13px] font-medium ${styles[variant]}`}
    >
      {children}
    </span>
  );
}
