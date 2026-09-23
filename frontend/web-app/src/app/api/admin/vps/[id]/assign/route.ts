import { NextResponse } from "next/server";
import { routeApiCall } from "@/lib/api";
import { getSelectedOrg } from "@/lib/session";
import type { VPSInstance } from "@/lib/types";

type Params = { params: { id: string } };

export async function PUT(request: Request, { params }: Params) {
  const org = getSelectedOrg();
  if (!org) {
    return NextResponse.json(
      { error: { message: "Selecciona una organización primero" } },
      { status: 400 },
    );
  }
  const body = await request.json().catch(() => ({}));
  const res = await routeApiCall<VPSInstance>({
    method: "PUT",
    path: `/api/v1/admin/vps/${params.id}/assign`,
    tenantId: org,
    body,
  });
  return NextResponse.json(
    { data: res.data, error: res.error },
    { status: res.error ? res.status : 200 },
  );
}
