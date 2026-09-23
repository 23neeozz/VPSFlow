export function MetricCard({
  label,
  value,
  hint,
  icon,
}: {
  label: string;
  value: string | number;
  hint?: string;
  icon?: React.ReactNode;
}) {
  return (
    <div className="rounded-card border border-bc-border bg-bc-card p-5 shadow-card transition-colors duration-200 hover:border-bc-border-hover hover:bg-bc-card-hover">
      <div className="flex items-start justify-between gap-3">
        <p className="text-[13px] font-medium text-bc-text-secondary">{label}</p>
        {icon && (
          <div className="text-bc-text-tertiary opacity-80">{icon}</div>
        )}
      </div>
      <p className="mt-3 text-2xl font-semibold tracking-tight text-bc-text-primary">
        {value}
      </p>
      {hint && (
        <p className="mt-1 text-[13px] text-bc-text-tertiary">{hint}</p>
      )}
    </div>
  );
}

export function EmptyState({
  title,
  description,
  action,
}: {
  title: string;
  description: string;
  action?: React.ReactNode;
}) {
  return (
    <div className="flex flex-col items-center justify-center rounded-card border border-dashed border-bc-border bg-bc-card/50 px-8 py-16 text-center">
      <h3 className="text-lg font-semibold text-bc-text-primary">{title}</h3>
      <p className="mt-2 max-w-sm text-[15px] text-bc-text-secondary">
        {description}
      </p>
      {action && <div className="mt-6">{action}</div>}
    </div>
  );
}

export function PageHeader({
  title,
  description,
  action,
}: {
  title: string;
  description?: string;
  action?: React.ReactNode;
}) {
  return (
    <div className="mb-8 flex flex-wrap items-end justify-between gap-4">
      <div>
        <h1 className="text-2xl font-bold tracking-tight text-bc-text-primary md:text-[28px]">
          {title}
        </h1>
        {description && (
          <p className="mt-2 max-w-2xl text-[15px] text-bc-text-secondary">
            {description}
          </p>
        )}
      </div>
      {action}
    </div>
  );
}
