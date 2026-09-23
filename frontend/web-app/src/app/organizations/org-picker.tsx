"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { ArrowRight, Building2, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input, Label } from "@/components/ui/input";
import type { Organization } from "@/lib/types";

export default function OrgPicker({
  redirectTo = "/app",
}: {
  redirectTo?: string;
}) {
  const router = useRouter();
  const [orgs, setOrgs] = useState<Organization[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [newName, setNewName] = useState("");
  const [creating, setCreating] = useState(false);

  async function load() {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch("/api/organizations");
      const data = await res.json();
      if (!res.ok) {
        setError(data?.error?.message || "No se pudieron cargar las organizaciones");
        return;
      }
      setOrgs(data.data || []);
    } catch {
      setError("Error de red al cargar organizaciones");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    load();
  }, []);

  async function select(orgId: string) {
    const res = await fetch("/api/organizations/select", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ org_id: orgId }),
    });
    if (res.ok) {
      router.replace(redirectTo);
      router.refresh();
    }
  }

  async function create(e: React.FormEvent) {
    e.preventDefault();
    if (!newName.trim()) return;
    setCreating(true);
    setError(null);
    try {
      const res = await fetch("/api/organizations", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name: newName.trim() }),
      });
      const data = await res.json();
      if (!res.ok) {
        setError(data?.error?.message || "No se pudo crear la organización");
        return;
      }
      setNewName("");
      if (data.data?.org_id) {
        await select(data.data.org_id);
      } else {
        await load();
      }
    } finally {
      setCreating(false);
    }
  }

  return (
    <div className="space-y-6">
      {error && (
        <div className="rounded-input border border-bc-danger/30 bg-bc-danger/10 px-4 py-3 text-[13px] text-bc-danger">
          {error}
        </div>
      )}

      {loading ? (
        <p className="text-[15px] text-bc-text-tertiary">Cargando...</p>
      ) : orgs.length === 0 ? (
        <Card>
          <p className="text-[15px] text-bc-text-secondary">
            Aún no tienes organizaciones. Crea la primera abajo.
          </p>
        </Card>
      ) : (
        <div className="space-y-3">
          {orgs.map((org) => (
            <button
              key={org.org_id}
              type="button"
              onClick={() => select(org.org_id)}
              className="flex w-full items-center justify-between rounded-card border border-bc-border bg-bc-card p-5 text-left shadow-card transition-colors duration-200 hover:border-bc-border-hover hover:bg-bc-card-hover"
            >
              <div className="flex items-center gap-4">
                <div className="flex h-11 w-11 items-center justify-center rounded-input border border-bc-border bg-bc-bg-secondary text-bc-primary-light">
                  <Building2 className="h-5 w-5" strokeWidth={1.75} />
                </div>
                <div>
                  <p className="font-semibold text-bc-text-primary">{org.name}</p>
                  <p className="mt-0.5 text-[13px] text-bc-text-tertiary">
                    {org.org_id}
                  </p>
                </div>
              </div>
              <ArrowRight className="h-5 w-5 text-bc-text-tertiary" />
            </button>
          ))}
        </div>
      )}

      <Card>
        <form onSubmit={create} className="space-y-4">
          <div>
            <Label>Nueva organización</Label>
            <Input
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              placeholder="Mi empresa"
            />
          </div>
          <Button type="submit" disabled={creating}>
            <Plus className="h-4 w-4" strokeWidth={1.75} />
            {creating ? "Creando..." : "Crear organización"}
          </Button>
        </form>
      </Card>
    </div>
  );
}
