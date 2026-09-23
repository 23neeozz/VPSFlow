import { redirect } from "next/navigation";
import { getSelectedOrg } from "@/lib/session";
import { UserShell } from "@/components/layout/user-shell";

export default function UserPanelLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const org = getSelectedOrg();
  if (!org) {
    redirect("/organizations?next=/app");
  }
  return <UserShell org={org}>{children}</UserShell>;
}
