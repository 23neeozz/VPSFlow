"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { Plus, Trash2, UserPlus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardHeader } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { EmptyState } from "@/components/ui/page";
import { StatusBadge } from "@/components/ui/status-badge";
import { formatRelative } from "@/lib/format";
import type { VPSInstance } from "@/lib/types";
import CreateVPSModal from "./create-vps-modal";

const ACTIVE = new Set(["pending", "provisioning", "starting", "stopping", "deleting"]);

export default function AdminVPSPanel() {
  const [items, setItems] = useState<VPSInstance[]>([]);
  const [loaded, setLoaded] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [busy, setBusy] = useState<Record<string, boolean>>({});
  const [assignEmail, setAssignEmail] = useState<Record<string, string>>({});

  const load = useCallback(async () => {
    try {
      const res = await fetch("/api/admin/vps");
      const data = await res.json();
      if (!res.ok) {
        setError(
          data?.error?.message ||
            "No tienes permisos de admin o no se pudieron cargar las VPS",
        );
        return;
      }
      setError(null);
      setItems(data.data || []);
    } catch {
      setError("Error de red al cargar las VPS");
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

  async function remove(id: string, name: string) {
    if (!confirm(`¿Eliminar la VPS "${name}"?`)) return;
    setBusy((b) => ({ ...b, [id]: true }));
    try {
      await fetch(`/api/admin/vps/${id}`, { method: "DELETE" });
      await load();
    } finally {
      setBusy((b) => ({ ...b, [id]: false }));
    }
  }

  async function assign(id: string) {
    const email = assignEmail[id]?.trim();
    if (!email) return;
    setBusy((b) => ({ ...b, [id]: true }));
    try {
      const res = await fetch(`/api/admin/vps/${id}/assign`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ owner_email: email }),
      });
      const data = await res.json();
      if (!res.ok) {
        setError(data?.error?.message || "No se pudo reasignar");
        return;
      }
      setAssignEmail((s) => ({ ...s, [id]: "" }));
      await load();
    } finally {
      setBusy((b) => ({ ...b, [id]: false }));
    }
  }

  return (
    <section>
      <CardHeader
        title="VPS de clientes"
        description="Provisiona máquinas virtuales y asígnalas a los usuarios de la organización."
        action={
          <Button onClick={() => setShowCreate(true)}>
            <Plus className="h-4 w-4" strokeWidth={1.75} />
            Crear VPS
          </Button>
        }
      />

      {error && (
        <div className="mb-4 rounded-input border border-bc-danger/30 bg-bc-danger/10 px-4 py-3 text-[13px] text-bc-danger">
          {error}
        </div>
      )}

      {!loaded ? (
        <div className="h-48 animate-pulse rounded-card border border-bc-border bg-bc-card" />
      ) : items.length === 0 ? (
        <EmptyState
          title="Sin VPS provisionadas"
          description="Crea la primera VPS para un cliente de tu organización."
          action={
            <Button onClick={() => setShowCreate(true)}>
              <Plus className="h-4 w-4" strokeWidth={1.75} />
              Crear VPS
            </Button>
          }
        />
      ) : (
        <Card className="!p-0 overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full min-w-[720px] text-left text-[15px]">
              <thead>
                <tr className="border-b border-bc-border bg-bc-bg-secondary/50 text-[13px] font-medium uppercase tracking-wider text-bc-text-tertiary">
                  <th className="px-6 py-4">VPS</th>
                  <th className="px-6 py-4">Cliente</th>
                  <th className="px-6 py-4">Estado</th>
                  <th className="px-6 py-4">Recursos</th>
                  <th className="px-6 py-4 text-right">Acciones</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-bc-border">
                {items.map((vps) => (
                  <tr
                    key={vps.id}
                    className="transition-colors hover:bg-bc-card-hover/50"
                  >
                    <td className="px-6 py-4">
                      <Link
                        href={`/admin/vps/${vps.id}`}
                        className="block hover:text-bc-text-primary"
                      >
                        <p className="font-medium text-bc-text-primary">
                          {vps.name}
                        </p>
                        <p className="text-[13px] text-bc-text-tertiary">{vps.id}</p>
                      </Link>
                    </td>
                    <td className="px-6 py-4">
                      <p className="text-bc-text-secondary">
                        {vps.owner_email || vps.owner_user_id}
                      </p>
                      <div className="mt-2 flex max-w-xs gap-2">
                        <Input
                          className="!h-9 !text-[13px]"
                          placeholder="reasignar@email.com"
                          value={assignEmail[vps.id] || ""}
                          onChange={(e) =>
                            setAssignEmail((s) => ({
                              ...s,
                              [vps.id]: e.target.value,
                            }))
                          }
                        />
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => assign(vps.id)}
                          disabled={busy[vps.id]}
                        >
                          <UserPlus className="h-4 w-4" strokeWidth={1.75} />
                        </Button>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <StatusBadge status={vps.status} />
                    </td>
                    <td className="px-6 py-4 text-[13px] text-bc-text-secondary">
                      {vps.vcpus}c · {vps.memory_mb}MB · {vps.disk_gb}GB
                      <br />
                      <span className="text-bc-text-tertiary">
                        {formatRelative(vps.updated_at)}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-right">
                      <Button
                        variant="danger"
                        size="sm"
                        onClick={() => remove(vps.id, vps.name)}
                        disabled={busy[vps.id]}
                      >
                        <Trash2 className="h-4 w-4" strokeWidth={1.75} />
                        Eliminar
                      </Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      )}

      {showCreate && (
        <CreateVPSModal
          onClose={() => setShowCreate(false)}
          onCreated={() => {
            setShowCreate(false);
            load();
          }}
        />
      )}
    </section>
  );
}
