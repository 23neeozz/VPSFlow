"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { ChevronRight, Monitor } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { EmptyState } from "@/components/ui/page";
import { StatusBadge } from "@/components/ui/status-badge";
import { formatRelative } from "@/lib/format";
import type { VPSInstance } from "@/lib/types";

const ACTIVE = new Set(["pending", "provisioning", "starting", "stopping", "deleting"]);

export default function VPSPanel() {
  const [items, setItems] = useState<VPSInstance[]>([]);
  const [loaded, setLoaded] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    try {
      const res = await fetch("/api/vps");
      const data = await res.json();
      if (!res.ok) {
        setError(data?.error?.message || "No se pudieron cargar tus VPS");
        return;
      }
      setError(null);
      setItems(data.data || []);
    } catch {
      setError("Error de red al cargar tus VPS");
    } finally {
      setLoaded(true);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  useEffect(() => {
    const anyActive = items.some((v) => ACTIVE.has(v.status));
    const t = setInterval(load, anyActive ? 2500 : 8000);
    return () => clearInterval(t);
  }, [items, load]);

  if (!loaded) {
    return (
      <div className="space-y-4">
        {[1, 2].map((i) => (
          <div
            key={i}
            className="h-28 animate-pulse rounded-card border border-bc-border bg-bc-card"
          />
        ))}
      </div>
    );
  }

  if (error && items.length === 0) {
    return (
      <EmptyState
        title="No se pudieron cargar las VPS"
        description={error}
        action={
          <Button variant="secondary" onClick={load}>
            Reintentar
          </Button>
        }
      />
    );
  }

  if (items.length === 0) {
    return (
      <EmptyState
        title="Sin VPS asignadas"
        description="Tu administrador aún no te ha asignado ninguna máquina virtual. Contacta con tu proveedor de hosting."
      />
    );
  }

  return (
    <div className="space-y-4">
      {error && (
        <div className="rounded-input border border-bc-danger/30 bg-bc-danger/10 px-4 py-3 text-[13px] text-bc-danger">
          {error}
        </div>
      )}

      {items.map((vps) => (
        <Link key={vps.id} href={`/app/vps/${vps.id}`} className="block">
          <Card hover className="!p-0 transition-all">
            <div className="flex items-center justify-between gap-4 p-6">
              <div className="flex items-start gap-4">
                <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-input border border-bc-border bg-bc-bg-secondary text-bc-primary-light">
                  <Monitor className="h-5 w-5" strokeWidth={1.75} />
                </div>
                <div>
                  <div className="flex flex-wrap items-center gap-3">
                    <h3 className="text-lg font-semibold text-bc-text-primary">
                      {vps.name}
                    </h3>
                    <StatusBadge status={vps.status} />
                  </div>
                  <p className="mt-2 text-[13px] text-bc-text-secondary">
                    {vps.vcpus} vCPU · {vps.memory_mb} MB RAM · {vps.disk_gb} GB
                    <span className="mx-2 text-bc-border">·</span>
                    Actualizada {formatRelative(vps.updated_at)}
                  </p>
                  {vps.error_message && (
                    <p className="mt-2 text-[13px] text-bc-danger line-clamp-1">
                      {vps.error_message}
                    </p>
                  )}
                </div>
              </div>
              <ChevronRight className="h-5 w-5 shrink-0 text-bc-text-tertiary" strokeWidth={1.75} />
            </div>
          </Card>
        </Link>
      ))}
    </div>
  );
}
