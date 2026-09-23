import { NextResponse } from "next/server";
import { routeApiCall } from "@/lib/api";
import { getSelectedOrg } from "@/lib/session";
import type { VPSInstance } from "@/lib/types";

type Params = { params: { id: string } };

export async function GET(_request: Request, { params }: Params) {
  const org = getSelectedOrg();
  if (!org) {
    return NextResponse.json(
      { error: { message: "Selecciona una organización primero" } },
      { status: 400 },
    );
  }
  const res = await routeApiCall<VPSInstance>({
    path: `/api/v1/admin/vps/${params.id}`,
    tenantId: org,
  });
  return NextResponse.json(
    { data: res.data, error: res.error },
    { status: res.error ? res.status : 200 },
  );
}

export async function DELETE(_request: Request, { params }: Params) {
  const org = getSelectedOrg();
  if (!org) {
    return NextResponse.json(
      { error: { message: "Selecciona una organización primero" } },
      { status: 400 },
    );
  }
  const res = await routeApiCall({
    method: "DELETE",
    path: `/api/v1/admin/vps/${params.id}`,
    tenantId: org,
  });
  return NextResponse.json(
    { ok: !res.error, error: res.error },
    { status: res.error ? res.status : 200 },
  );
}
