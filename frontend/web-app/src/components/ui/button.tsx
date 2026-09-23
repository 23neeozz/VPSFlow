"use client";

import { forwardRef } from "react";

type Variant = "primary" | "secondary" | "ghost" | "danger";

const variants: Record<Variant, string> = {
  primary: "bg-bc-primary text-black hover:bg-bc-primary-hover",
  secondary:
    "border border-bc-border bg-transparent text-bc-text-primary hover:border-bc-border-hover hover:bg-bc-card",
  ghost:
    "bg-transparent text-bc-text-secondary hover:bg-bc-card hover:text-bc-text-primary",
  danger: "bg-bc-danger/90 text-white hover:bg-bc-danger",
};

const sizes: Record<string, string> = {
  sm: "h-9 px-4 text-[13px]",
  md: "h-11 px-5 text-[15px]",
  lg: "h-12 px-6 text-[15px] font-semibold",
};

export const Button = forwardRef<
  HTMLButtonElement,
  React.ButtonHTMLAttributes<HTMLButtonElement> & {
    variant?: Variant;
    size?: keyof typeof sizes;
  }
>(function Button(
  { children, variant = "primary", size = "md", className = "", ...props },
  ref,
) {
  return (
    <button
      ref={ref}
      className={`inline-flex items-center justify-center gap-2 rounded-btn font-medium transition-colors duration-200 ease-in-out disabled:cursor-not-allowed disabled:opacity-50 ${variants[variant]} ${sizes[size]} ${className}`}
      {...props}
    >
      {children}
    </button>
  );
});
