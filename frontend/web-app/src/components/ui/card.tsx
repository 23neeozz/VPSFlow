export function Card({
  children,
  className = "",
  hover = false,
}: {
  children: React.ReactNode;
  className?: string;
  hover?: boolean;
}) {
  return (
    <div
      className={`rounded-card border border-bc-border bg-bc-card p-6 shadow-card transition-colors duration-200 ${
        hover ? "hover:border-bc-border-hover hover:bg-bc-card-hover" : ""
      } ${className}`}
    >
      {children}
    </div>
  );
}

export function CardHeader({
  title,
  description,
  action,
}: {
  title: string;
  description?: string;
  action?: React.ReactNode;
}) {
  return (
    <div className="mb-6 flex flex-wrap items-start justify-between gap-4">
      <div>
        <h2 className="text-lg font-semibold tracking-tight text-bc-text-primary">
          {title}
        </h2>
        {description && (
          <p className="mt-1 text-[13px] text-bc-text-secondary">
            {description}
          </p>
        )}
      </div>
      {action}
    </div>
  );
}
