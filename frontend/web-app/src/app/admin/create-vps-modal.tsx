"use client";

import { useState } from "react";
import { X } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";
import { Button } from "@/components/ui/button";
import { Input, Label } from "@/components/ui/input";

type Props = {
  onClose: () => void;
  onCreated: () => void;
};

const PRESETS = [
  { label: "Small", vcpus: 1, memory_mb: 1024, disk_gb: 10 },
  { label: "Medium", vcpus: 2, memory_mb: 2048, disk_gb: 20 },
  { label: "Large", vcpus: 4, memory_mb: 8192, disk_gb: 40 },
];

export default function CreateVPSModal({ onClose, onCreated }: Props) {
  const [name, setName] = useState("");
  const [ownerEmail, setOwnerEmail] = useState("");
  const [vcpus, setVcpus] = useState(2);
  const [memoryMb, setMemoryMb] = useState(2048);
  const [diskGb, setDiskGb] = useState(20);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError(null);
    try {
      const res = await fetch("/api/admin/vps", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: name.trim(),
          owner_email: ownerEmail.trim(),
          vcpus,
          memory_mb: memoryMb,
          disk_gb: diskGb,
        }),
      });
      const data = await res.json();
      if (!res.ok) {
        setError(data?.error?.message || "No se pudo crear la VPS");
        return;
      }
      onCreated();
    } catch {
      setError("Error de red al crear la VPS");
    } finally {
      setSaving(false);
    }
  }

  return (
    <AnimatePresence>
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4 backdrop-blur-sm"
        onClick={onClose}
      >
        <motion.div
          initial={{ opacity: 0, y: 8 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: 8 }}
          transition={{ duration: 0.25, ease: "easeInOut" }}
          className="w-full max-w-lg rounded-modal border border-bc-border bg-bc-card p-8 shadow-card"
          onClick={(e) => e.stopPropagation()}
        >
          <div className="mb-6 flex items-start justify-between">
            <div>
              <h3 className="text-xl font-bold text-bc-text-primary">
                Nueva VPS
              </h3>
              <p className="mt-1 text-[13px] text-bc-text-secondary">
                Provisiona y asigna a un cliente de la organización.
              </p>
            </div>
            <button
              type="button"
              onClick={onClose}
              className="rounded-input p-2 text-bc-text-tertiary hover:bg-bc-bg-secondary hover:text-bc-text-primary"
            >
              <X className="h-5 w-5" strokeWidth={1.75} />
            </button>
          </div>

          <form onSubmit={submit} className="space-y-5">
            <div>
              <Label>Nombre</Label>
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="cliente-web-01"
                required
              />
            </div>
            <div>
              <Label>Email del cliente</Label>
              <Input
                type="email"
                value={ownerEmail}
                onChange={(e) => setOwnerEmail(e.target.value)}
                placeholder="cliente@empresa.com"
                required
              />
            </div>

            <div>
              <Label>Perfil de recursos</Label>
              <div className="mt-2 grid grid-cols-3 gap-2">
                {PRESETS.map((p) => (
                  <button
                    key={p.label}
                    type="button"
                    onClick={() => {
                      setVcpus(p.vcpus);
                      setMemoryMb(p.memory_mb);
                      setDiskGb(p.disk_gb);
                    }}
                    className="rounded-input border border-bc-border bg-bc-bg-secondary/60 px-3 py-3 text-center transition-colors hover:border-bc-border-hover hover:bg-bc-card"
                  >
                    <span className="block text-[13px] font-semibold text-bc-text-primary">
                      {p.label}
                    </span>
                    <span className="mt-1 block text-[11px] text-bc-text-tertiary">
                      {p.vcpus}c · {p.memory_mb / 1024}G
                    </span>
                  </button>
                ))}
              </div>
            </div>

            <div className="grid grid-cols-3 gap-3">
              <div>
                <Label>vCPU</Label>
                <Input
                  type="number"
                  min={1}
                  value={vcpus}
                  onChange={(e) => setVcpus(Number(e.target.value))}
                />
              </div>
              <div>
                <Label>RAM MB</Label>
                <Input
                  type="number"
                  min={512}
                  step={512}
                  value={memoryMb}
                  onChange={(e) => setMemoryMb(Number(e.target.value))}
                />
              </div>
              <div>
                <Label>Disco GB</Label>
                <Input
                  type="number"
                  min={1}
                  value={diskGb}
                  onChange={(e) => setDiskGb(Number(e.target.value))}
                />
              </div>
            </div>

            {error && (
              <div className="rounded-input border border-bc-danger/30 bg-bc-danger/10 px-4 py-3 text-[13px] text-bc-danger">
                {error}
              </div>
            )}

            <div className="flex justify-end gap-3 pt-2">
              <Button type="button" variant="ghost" onClick={onClose}>
                Cancelar
              </Button>
              <Button type="submit" disabled={saving}>
                {saving ? "Creando..." : "Crear y asignar"}
              </Button>
            </div>
          </form>
        </motion.div>
      </motion.div>
    </AnimatePresence>
  );
}
