"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import {
  ArrowLeft,
  Cpu,
  HardDrive,
  MemoryStick,
  Monitor,
  Power,
  RefreshCw,
  Square,
  Terminal,
  Trash2,
  UserPlus,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { MetricCard } from "@/components/ui/page";
import { StatusBadge } from "@/components/ui/status-badge";
import { formatRelative } from "@/lib/format";
import type { ConsoleSession, VPSInstance } from "@/lib/types";

const ACTIVE = new Set(["pending", "provisioning", "starting", "stopping", "deleting"]);

const STATUS_HINT: Record<string, string> = {
  running: "La máquina está encendida y operativa.",
  stopped: "La máquina está apagada.",
  pending: "Esperando inicio del aprovisionamiento.",
  provisioning: "Creando la máquina virtual en el hipervisor.",
  starting: "Arrancando la máquina virtual.",
  stopping: "Apagando la máquina virtual.",
  deleting: "Eliminando la máquina virtual.",
  error: "Se produjo un error. Revisa el mensaje inferior.",
};

type VPSDetailPanelProps = {
  vpsId: string;
  backHref: string;
  mode: "user" | "admin";
};

export function VPSDetailPanel({ vpsId, backHref, mode }: VPSDetailPanelProps) {
  const router = useRouter();
  const [vps, setVps] = useState<VPSInstance | null>(null);
  const [loaded, setLoaded] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [assignEmail, setAssignEmail] = useState("");
  const [consoleInfo, setConsoleInfo] = useState<ConsoleSession | null>(null);

  const apiBase = mode === "admin" ? "/api/admin/vps" : "/api/vps";
  const actionBase = "/api/vps";

  const load = useCallback(async () => {
    try {
      const res = await fetch(`${apiBase}/${vpsId}`);
      const data = await res.json();
      if (!res.ok) {
        setError(data?.error?.message || "No se pudo cargar la VPS");
        setVps(null);
        return;
      }
      setError(null);
      setVps(data.data);
    } catch {
      setError("Error de red al cargar la VPS");
      setVps(null);
    } finally {
      setLoaded(true);
    }
  }, [apiBase, vpsId]);

  useEffect(() => {
    load();
  }, [load]);

  useEffect(() => {
    if (!vps || !ACTIVE.has(vps.status)) return;
    const t = setInterval(load, 2500);
    return () => clearInterval(t);
  }, [vps, load]);

  async function act(action: "start" | "stop" | "console") {
    setBusy(true);
    setError(null);
    try {
      const res = await fetch(`${actionBase}/${vpsId}/${action}`, { method: "POST" });
      const data = await res.json();
      if (!res.ok) {
        setError(data?.error?.message || "No se pudo completar la acción");
        return;
      }
      if (action === "console" && data.data?.proxy_url) {
        setConsoleInfo(data.data);
      }
      await load();
    } finally {
      setBusy(false);
    }
  }

  async function restart() {
    if (!confirm("¿Reiniciar esta VPS? Se apagará y volverá a encender.")) return;
    setBusy(true);
    setError(null);
    try {
      const stopRes = await fetch(`${actionBase}/${vpsId}/stop`, { method: "POST" });
      if (!stopRes.ok) {
        const data = await stopRes.json();
        setError(data?.error?.message || "No se pudo detener la VPS");
        return;
      }
      await new Promise((r) => setTimeout(r, 1500));
      const startRes = await fetch(`${actionBase}/${vpsId}/start`, { method: "POST" });
      const data = await startRes.json();
      if (!startRes.ok) {
        setError(data?.error?.message || "No se pudo iniciar la VPS");
        return;
      }
      await load();
    } finally {
      setBusy(false);
    }
  }

  async function assign() {
    const email = assignEmail.trim();
    if (!email) return;
    setBusy(true);
    setError(null);
    try {
      const res = await fetch(`${apiBase}/${vpsId}/assign`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ owner_email: email }),
      });
      const data = await res.json();
      if (!res.ok) {
        setError(data?.error?.message || "No se pudo reasignar");
        return;
      }
      setAssignEmail("");
      await load();
    } finally {
      setBusy(false);
    }
  }

  async function remove() {
    if (!vps || !confirm(`¿Eliminar la VPS "${vps.name}"?`)) return;
    setBusy(true);
    try {
      const res = await fetch(`${apiBase}/${vpsId}`, { method: "DELETE" });
      if (!res.ok) {
        const data = await res.json();
        setError(data?.error?.message || "No se pudo eliminar");
        return;
      }
      router.push(backHref);
      router.refresh();
    } finally {
      setBusy(false);
    }
  }

  if (!loaded) {
    return (
      <div className="space-y-4">
        <div className="h-10 w-40 animate-pulse rounded-input bg-bc-card" />
        <div className="h-64 animate-pulse rounded-card border border-bc-border bg-bc-card" />
      </div>
    );
  }

  if (!vps) {
    return (
      <div className="space-y-4">
        <Link
          href={backHref}
          className="inline-flex items-center gap-2 text-[13px] text-bc-text-secondary transition-colors hover:text-bc-text-primary"
        >
          <ArrowLeft className="h-4 w-4" strokeWidth={1.75} />
          Volver al listado
        </Link>
        <Card>
          <p className="text-bc-text-primary">VPS no encontrada</p>
          <p className="mt-2 text-[13px] text-bc-text-secondary">
            {error || "No existe o no tienes acceso."}
          </p>
        </Card>
      </div>
    );
  }

  const canStart = vps.status === "stopped";
  const canStop = vps.status === "running";
  const canRestart = vps.status === "running" || vps.status === "stopped";
  const canConsole = vps.status === "running";
  const isTransition = ACTIVE.has(vps.status);

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <Link
          href={backHref}
          className="inline-flex items-center gap-2 text-[13px] text-bc-text-secondary transition-colors hover:text-bc-text-primary"
        >
          <ArrowLeft className="h-4 w-4" strokeWidth={1.75} />
          Volver al listado
        </Link>
        {mode === "admin" && (
          <Button variant="danger" size="sm" onClick={remove} disabled={busy}>
            <Trash2 className="h-4 w-4" strokeWidth={1.75} />
            Eliminar VPS
          </Button>
        )}
      </div>

      {error && (
        <div className="rounded-input border border-bc-danger/30 bg-bc-danger/10 px-4 py-3 text-[13px] text-bc-danger">
          {error}
        </div>
      )}

      <div className="grid gap-6 lg:grid-cols-[1fr_320px]">
        <div className="space-y-6">
          <Card className="!p-0 overflow-hidden">
            <div className="border-b border-bc-border bg-bc-bg-secondary/40 px-6 py-5">
              <div className="flex flex-wrap items-start justify-between gap-4">
                <div className="flex items-start gap-4">
                  <div className="flex h-14 w-14 shrink-0 items-center justify-center rounded-input border border-bc-border bg-bc-card">
                    <Monitor className="h-6 w-6 text-bc-primary" strokeWidth={1.75} />
                  </div>
                  <div>
                    <h1 className="text-2xl font-bold tracking-tight text-bc-text-primary">
                      {vps.name}
                    </h1>
                    <p className="mt-1 font-mono text-[13px] text-bc-text-tertiary">{vps.id}</p>
                    <div className="mt-3">
                      <StatusBadge status={vps.status} />
                    </div>
                  </div>
                </div>
              </div>
              <p className="mt-4 text-[14px] text-bc-text-secondary">
                {STATUS_HINT[vps.status] || "Estado de la máquina virtual."}
              </p>
              {vps.error_message && (
                <p className="mt-3 rounded-input border border-bc-danger/30 bg-bc-danger/10 px-4 py-3 text-[13px] text-bc-danger">
                  {vps.error_message}
                </p>
              )}
            </div>

            <div className="p-6">
              <p className="mb-4 text-[13px] font-medium uppercase tracking-wider text-bc-text-tertiary">
                Control de energía
              </p>
              <div className="flex flex-wrap gap-3">
                <Button
                  onClick={() => act("start")}
                  disabled={busy || !canStart || isTransition}
                >
                  <Power className="h-4 w-4" strokeWidth={1.75} />
                  Encender
                </Button>
                <Button
                  variant="secondary"
                  onClick={() => act("stop")}
                  disabled={busy || !canStop || isTransition}
                >
                  <Square className="h-4 w-4" strokeWidth={1.75} />
                  Apagar
                </Button>
                <Button
                  variant="secondary"
                  onClick={restart}
                  disabled={busy || !canRestart || isTransition}
                >
                  <RefreshCw className="h-4 w-4" strokeWidth={1.75} />
                  Reiniciar
                </Button>
                <Button
                  variant="secondary"
                  onClick={() => act("console")}
                  disabled={busy || !canConsole}
                >
                  <Terminal className="h-4 w-4" strokeWidth={1.75} />
                  Consola
                </Button>
              </div>
            </div>
          </Card>

          <div className="grid gap-4 sm:grid-cols-3">
            <MetricCard
              label="vCPU"
              value={vps.vcpus}
              icon={<Cpu className="h-4 w-4" strokeWidth={1.75} />}
            />
            <MetricCard
              label="Memoria"
              value={`${vps.memory_mb} MB`}
              icon={<MemoryStick className="h-4 w-4" strokeWidth={1.75} />}
            />
            <MetricCard
              label="Disco"
              value={`${vps.disk_gb} GB`}
              icon={<HardDrive className="h-4 w-4" strokeWidth={1.75} />}
            />
          </div>

          {consoleInfo && (
            <Card className="border-bc-primary/30 bg-bc-primary/5">
              <h3 className="font-semibold text-bc-text-primary">Sesión de consola</h3>
              <p className="mt-2 text-[13px] text-bc-text-secondary">
                Token generado. El visor noVNC se integrará en la siguiente fase.
              </p>
              <code className="mt-3 block break-all rounded-input border border-bc-border bg-bc-bg px-4 py-3 text-[13px] text-bc-primary-light">
                {consoleInfo.proxy_url}
              </code>
              <Button
                variant="ghost"
                size="sm"
                className="mt-4"
                onClick={() => setConsoleInfo(null)}
              >
                Cerrar
              </Button>
            </Card>
          )}
        </div>

        <div className="space-y-4">
          <Card>
            <h3 className="text-[13px] font-medium uppercase tracking-wider text-bc-text-tertiary">
              Información
            </h3>
            <dl className="mt-4 space-y-4 text-[14px]">
              <div>
                <dt className="text-bc-text-tertiary">VM ID</dt>
                <dd className="mt-1 break-all font-mono text-[13px] text-bc-text-secondary">
                  {vps.vm_id}
                </dd>
              </div>
              {mode === "admin" && (
                <div>
                  <dt className="text-bc-text-tertiary">Cliente</dt>
                  <dd className="mt-1 text-bc-text-primary">
                    {vps.owner_email || vps.owner_user_id}
                  </dd>
                </div>
              )}
              <div>
                <dt className="text-bc-text-tertiary">Creada</dt>
                <dd className="mt-1 text-bc-text-secondary">
                  {formatRelative(vps.created_at)}
                </dd>
              </div>
              <div>
                <dt className="text-bc-text-tertiary">Última actualización</dt>
                <dd className="mt-1 text-bc-text-secondary">
                  {formatRelative(vps.updated_at)}
                </dd>
              </div>
            </dl>
          </Card>

          {mode === "admin" && (
            <Card>
              <h3 className="text-[13px] font-medium uppercase tracking-wider text-bc-text-tertiary">
                Reasignar cliente
              </h3>
              <p className="mt-2 text-[13px] text-bc-text-secondary">
                El usuario debe ser miembro de la organización.
              </p>
              <div className="mt-4 flex gap-2">
                <Input
                  className="!h-10 !text-[13px]"
                  placeholder="cliente@email.com"
                  value={assignEmail}
                  onChange={(e) => setAssignEmail(e.target.value)}
                />
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={assign}
                  disabled={busy || !assignEmail.trim()}
                >
                  <UserPlus className="h-4 w-4" strokeWidth={1.75} />
                </Button>
              </div>
            </Card>
          )}
        </div>
      </div>
    </div>
  );
}
