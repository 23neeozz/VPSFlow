"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import {
  Building2,
  LayoutDashboard,
  LogOut,
  Server,
  Settings2,
  Shield,
  Users,
} from "lucide-react";
import { Logo, LogoMark } from "@/components/brand/logo";
import { Button } from "@/components/ui/button";

const navItems = [
  {
    href: "/admin",
    label: "VPS clientes",
    icon: Users,
    match: (p: string) => p === "/admin",
  },
  {
    href: "/admin/infra",
    label: "Infraestructura",
    icon: Server,
    match: (p: string) => p.startsWith("/admin/infra"),
  },
  {
    href: "/admin/organizations",
    label: "Organización",
    icon: Building2,
    match: (p: string) => p.startsWith("/admin/organizations"),
  },
];

export function AdminShell({
  children,
  org,
}: {
  children: React.ReactNode;
  org: string;
}) {
  const pathname = usePathname();
  const router = useRouter();

  async function logout() {
    await fetch("/api/auth/logout", { method: "POST" });
    router.replace("/login");
  }

  return (
    <div className="flex min-h-screen">
      <aside className="fixed inset-y-0 left-0 z-40 flex w-[260px] flex-col border-r border-bc-border bg-bc-sidebar">
        <div className="flex h-20 items-center gap-3 border-b border-bc-border px-6">
          <LogoMark />
          <div className="min-w-0">
            <p className="truncate text-sm font-semibold text-bc-text-primary">
              VPSFlow
            </p>
            <p className="truncate text-[11px] text-bc-text-tertiary">
              Panel de administración
            </p>
          </div>
        </div>

        <div className="border-b border-bc-border px-4 py-3">
          <span className="inline-flex items-center gap-1.5 rounded-full border border-bc-primary/30 bg-bc-primary/10 px-3 py-1 text-[12px] font-medium text-bc-primary-light">
            <Shield className="h-3.5 w-3.5" strokeWidth={1.75} />
            Admin
          </span>
        </div>

        <nav className="flex-1 space-y-1 p-4">
          {navItems.map((item) => {
            const active = item.match(pathname);
            const Icon = item.icon;
            return (
              <Link
                key={item.href}
                href={item.href}
                className={`flex items-center gap-3 rounded-input px-4 py-3 text-[15px] font-medium transition-all duration-250 ${
                  active
                    ? "bg-bc-primary/10 text-bc-primary-light"
                    : "text-bc-text-secondary hover:bg-bc-card hover:text-bc-text-primary"
                }`}
              >
                <Icon className="h-[18px] w-[18px]" strokeWidth={1.75} />
                {item.label}
              </Link>
            );
          })}
        </nav>

        <div className="border-t border-bc-border p-4">
          <div className="mb-3 rounded-input border border-bc-border bg-bc-card px-3 py-2.5">
            <p className="text-[11px] font-medium uppercase tracking-wider text-bc-text-tertiary">
              Organización
            </p>
            <p
              className="mt-1 truncate text-[13px] text-bc-text-secondary"
              title={org}
            >
              {org}
            </p>
          </div>
          <Button
            variant="ghost"
            size="sm"
            className="w-full justify-start"
            onClick={logout}
          >
            <LogOut className="h-4 w-4" strokeWidth={1.75} />
            Cerrar sesión
          </Button>
        </div>
      </aside>

      <div className="flex min-h-screen flex-1 flex-col pl-[260px]">
        <header className="sticky top-0 z-30 flex h-20 items-center justify-between border-b border-bc-border bg-bc-bg-secondary/80 px-8 backdrop-blur-xl">
          <Logo className="lg:hidden" />
          <div className="hidden items-center gap-2 text-bc-text-tertiary lg:flex">
            <LayoutDashboard className="h-4 w-4" strokeWidth={1.75} />
            <span className="text-[13px]">Área de administración</span>
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => router.push("/admin/organizations")}
            >
              <Settings2 className="h-4 w-4" strokeWidth={1.75} />
              Cambiar org
            </Button>
          </div>
        </header>
        <main className="flex-1 p-8">{children}</main>
      </div>
    </div>
  );
}
