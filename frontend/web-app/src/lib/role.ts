import { apiCall } from "@/lib/api";
import { getSelectedOrg } from "@/lib/session";

export async function getMembershipRole(): Promise<{
  role: string;
  isAdmin: boolean;
} | null> {
  const org = getSelectedOrg();
  if (!org) return null;

  const me = await apiCall<{ user_id: string }>({
    path: "/api/v1/protected/me",
  });
  if (!me.data?.user_id) return null;

  const members = await apiCall<{
    data: { user_id: string; role: string }[];
  }>({
    path: `/api/v1/organizations/${org}/members`,
  });

  const role =
    members.data?.data?.find((m) => m.user_id === me.data?.user_id)?.role ?? "";

  return {
    role,
    isAdmin: role === "owner" || role === "admin",
  };
}
