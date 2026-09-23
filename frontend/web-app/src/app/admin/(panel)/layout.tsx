import { redirect } from "next/navigation";
import { getSelectedOrg } from "@/lib/session";
import { getMembershipRole } from "@/lib/role";
import { AdminShell } from "@/components/layout/admin-shell";

export default async function AdminPanelLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const org = getSelectedOrg();
  if (!org) {
    redirect("/organizations?next=/admin");
  }

  const membership = await getMembershipRole();
  if (!membership?.isAdmin) {
    redirect("/app");
  }

  return <AdminShell org={org}>{children}</AdminShell>;
}
