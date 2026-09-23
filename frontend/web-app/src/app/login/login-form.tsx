"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input, Label } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";

type Mode = "login" | "register";

export default function LoginForm() {
  const [mode, setMode] = useState<Mode>("login");
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), 20000);
    try {
      const endpoint = mode === "login" ? "/api/auth/login" : "/api/auth/register";
      const payload =
        mode === "login"
          ? { email, password }
          : { email, password, name };
      const res = await fetch(endpoint, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
        signal: controller.signal,
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok || !data.ok) {
        setError(data?.error?.message || "No se pudo completar la operación");
        return;
      }
      window.location.href = "/";
    } catch (err) {
      const aborted = err instanceof Error && err.name === "AbortError";
      setError(
        aborted
          ? "Tiempo de espera agotado. Comprueba que el backend esté en marcha."
          : "Error de red. ¿Está el servidor en marcha?",
      );
    } finally {
      clearTimeout(timer);
      setLoading(false);
    }
  }

  return (
    <Card className="!p-0 overflow-hidden">
      <div className="flex border-b border-bc-border p-1.5">
        {(["login", "register"] as const).map((m) => (
          <button
            key={m}
            type="button"
            onClick={() => setMode(m)}
            className={`flex-1 rounded-input py-2.5 text-[15px] font-medium transition-all duration-250 ${
              mode === m
                ? "bg-bc-primary text-black"
                : "text-bc-text-tertiary hover:text-bc-text-secondary"
            }`}
          >
            {m === "login" ? "Iniciar sesión" : "Crear cuenta"}
          </button>
        ))}
      </div>

      <form onSubmit={submit} className="space-y-5 p-6">
        {mode === "register" && (
          <div>
            <Label>Nombre</Label>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Tu nombre"
              required
            />
          </div>
        )}
        <div>
          <Label>Email</Label>
          <Input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="tu@empresa.com"
            required
          />
        </div>
        <div>
          <Label>Contraseña</Label>
          <Input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="••••••••••••"
            required
            minLength={mode === "register" ? 12 : undefined}
          />
          {mode === "register" && (
            <p className="mt-2 text-[13px] text-bc-text-tertiary">
              Mínimo 12 caracteres con mayúscula, minúscula, número y símbolo.
            </p>
          )}
        </div>

        {error && (
          <div className="rounded-input border border-bc-danger/30 bg-bc-danger/10 px-4 py-3 text-[13px] text-bc-danger">
            {error}
          </div>
        )}

        <Button type="submit" className="w-full" size="lg" disabled={loading}>
          {loading
            ? "Procesando..."
            : mode === "login"
              ? "Entrar al panel"
              : "Crear cuenta"}
        </Button>

        <div className="flex flex-wrap gap-2 pt-1">
          <Badge variant="primary">Enterprise</Badge>
          <Badge>Multi-tenant</Badge>
          <Badge>Seguro</Badge>
        </div>
      </form>
    </Card>
  );
}
