import { redirect } from "next/navigation";
import { isAuthenticated, getSelectedOrg } from "@/lib/session";
import { getMembershipRole } from "@/lib/role";

export default async function Home({
  searchParams,
}: {
  searchParams: { next?: string };
}) {
  if (!isAuthenticated()) {
    redirect("/login");
  }
  if (!getSelectedOrg()) {
    const next = searchParams.next || "/app";
    redirect(`/organizations?next=${encodeURIComponent(next)}`);
  }

  const membership = await getMembershipRole();
  redirect(membership?.isAdmin ? "/admin" : "/app");
}
