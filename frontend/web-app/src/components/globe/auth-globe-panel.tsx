"use client";

import { RotatingGlobe } from "@/components/globe/rotating-globe";

const REGIONS = [
  "Madrid",
  "Barcelona",
  "París",
  "Frankfurt",
  "Amsterdam",
  "New York",
  "Miami",
  "São Paulo",
];

export function AuthGlobePanel() {
  return (
    <div className="relative hidden min-h-[520px] overflow-hidden rounded-card border border-bc-border bg-bc-card lg:flex lg:flex-col lg:items-center lg:justify-center">
      <div className="absolute inset-0 bg-gradient-to-b from-bc-primary/8 via-transparent to-bc-bg-secondary/40" />
      <div className="absolute left-1/2 top-1/2 h-64 w-64 -translate-x-1/2 -translate-y-1/2 rounded-full bg-bc-primary/15 blur-[80px]" />

      <div className="relative z-10 flex w-full flex-col items-center px-8 py-10">
        <p className="mb-2 text-[13px] font-medium uppercase tracking-[0.2em] text-bc-primary-light">
          Red global
        </p>
        <p className="mb-8 text-center text-[15px] text-bc-text-secondary">
          15 regiones · 10 Tbps · 99.99% disponibilidad
        </p>

        <RotatingGlobe size={400} />

        <div className="mt-8 flex flex-wrap justify-center gap-2">
          {REGIONS.map((city) => (
            <span
              key={city}
              className="rounded-full border border-bc-border bg-bc-bg-secondary/60 px-3 py-1 text-[12px] text-bc-text-tertiary"
            >
              <span className="mr-1.5 inline-block h-1.5 w-1.5 rounded-full bg-bc-primary" />
              {city}
            </span>
          ))}
        </div>
      </div>
    </div>
  );
}
