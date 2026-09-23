"use client";

import { useEffect, useState } from "react";
import { Cpu, HardDrive, MemoryStick, Server } from "lucide-react";
import { Card, CardHeader } from "@/components/ui/card";
import { MetricCard } from "@/components/ui/page";
import { StatusBadge } from "@/components/ui/status-badge";
import { bytesToGiB, formatRelative } from "@/lib/format";
import type { Hypervisor } from "@/lib/types";

export default function HypervisorPanel() {
  const [nodes, setNodes] = useState<Hypervisor[]>([]);
  const [loaded, setLoaded] = useState(false);

  async function load() {
    try {
      const res = await fetch("/api/hypervisors");
      const data = await res.json();
      if (res.ok) setNodes(data.data || []);
    } finally {
      setLoaded(true);
    }
  }

  useEffect(() => {
    load();
    const t = setInterval(load, 10000);
    return () => clearInterval(t);
  }, []);

  const online = nodes.filter((n) => n.status === "online").length;
  const totalVms = nodes.reduce((s, n) => s + n.capacity.running_vms, 0);

  return (
    <section>
      <CardHeader
        title="Infraestructura"
        description="Estado de los hipervisores y capacidad del cluster."
      />

      <div className="mb-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <MetricCard
          label="Nodos online"
          value={loaded ? `${online}/${nodes.length}` : "—"}
          icon={<Server className="h-4 w-4" strokeWidth={1.75} />}
        />
        <MetricCard
          label="VMs activas"
          value={loaded ? totalVms : "—"}
          icon={<Cpu className="h-4 w-4" strokeWidth={1.75} />}
        />
      </div>

      {!loaded ? (
        <div className="grid gap-4 sm:grid-cols-2">
          {[1, 2].map((i) => (
            <div
              key={i}
              className="h-40 animate-pulse rounded-card border border-bc-border bg-bc-card"
            />
          ))}
        </div>
      ) : nodes.length === 0 ? (
        <Card>
          <p className="text-[15px] text-bc-text-secondary">
            No hay hipervisores registrados. Arranca el hypervisor-agent.
          </p>
        </Card>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {nodes.map((n) => (
            <Card key={n.id} hover>
              <div className="mb-4 flex items-start justify-between">
                <div>
                  <p className="font-semibold text-bc-text-primary">{n.node_name}</p>
                  <p className="text-[13px] text-bc-text-tertiary">v{n.agent_version}</p>
                </div>
                <StatusBadge status={n.status} />
              </div>
              <dl className="space-y-2.5 text-[13px]">
                <div className="flex justify-between text-bc-text-secondary">
                  <dt className="flex items-center gap-2">
                    <Cpu className="h-3.5 w-3.5" strokeWidth={1.75} />
                    vCPU
                  </dt>
                  <dd className="font-medium text-bc-text-primary">
                    {n.capacity.allocated_cpu}/{n.capacity.cpu_cores}
                  </dd>
                </div>
                <div className="flex justify-between text-bc-text-secondary">
                  <dt className="flex items-center gap-2">
                    <MemoryStick className="h-3.5 w-3.5" strokeWidth={1.75} />
                    Memoria
                  </dt>
                  <dd className="font-medium text-bc-text-primary">
                    {bytesToGiB(n.capacity.allocated_memory_bytes)}/
                    {bytesToGiB(n.capacity.memory_bytes)} GiB
                  </dd>
                </div>
                <div className="flex justify-between text-bc-text-secondary">
                  <dt className="flex items-center gap-2">
                    <HardDrive className="h-3.5 w-3.5" strokeWidth={1.75} />
                    VMs
                  </dt>
                  <dd className="font-medium text-bc-text-primary">
                    {n.capacity.running_vms}
                  </dd>
                </div>
                <div className="flex justify-between border-t border-bc-border pt-2 text-bc-text-tertiary">
                  <dt>Heartbeat</dt>
                  <dd>{formatRelative(n.last_heartbeat_at)}</dd>
                </div>
              </dl>
            </Card>
          ))}
        </div>
      )}
    </section>
  );
}
