import { forwardRef } from "react";

export const Input = forwardRef<
  HTMLInputElement,
  React.InputHTMLAttributes<HTMLInputElement>
>(function Input({ className = "", ...props }, ref) {
  return (
    <input
      ref={ref}
      className={`h-11 w-full rounded-input border border-bc-border bg-bc-bg-secondary/80 px-4 text-[15px] text-bc-text-primary outline-none transition-colors duration-200 placeholder:text-bc-text-tertiary focus:border-bc-border-hover ${className}`}
      {...props}
    />
  );
});

export function Label({
  children,
  className = "",
}: {
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <label
      className={`mb-2 block text-[13px] font-medium text-bc-text-secondary ${className}`}
    >
      {children}
    </label>
  );
}
