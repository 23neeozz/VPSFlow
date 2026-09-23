import { NextResponse } from "next/server";
import { routeApiCall } from "@/lib/api";
import { getSelectedOrg } from "@/lib/session";

export async function GET() {
  const org = getSelectedOrg();
  if (!org) {
    return NextResponse.json(
      { error: { message: "Selecciona una organización primero" } },
      { status: 400 },
    );
  }

  const me = await routeApiCall<{ user_id: string }>({
    path: "/api/v1/protected/me",
  });
  if (!me.data?.user_id) {
    return NextResponse.json(
      { error: me.error ?? { message: "No autenticado" } },
      { status: me.status || 401 },
    );
  }

  const members = await routeApiCall<{
    data: { user_id: string; role: string }[];
  }>({
    path: `/api/v1/organizations/${org}/members`,
  });
  const role =
    members.data?.data?.find((m) => m.user_id === me.data?.user_id)?.role ?? "";
  const isAdmin = role === "owner" || role === "admin";

  return NextResponse.json({ role, is_admin: isAdmin });
}
