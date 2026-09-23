import { NextResponse } from "next/server";
import { routeApiCall } from "@/lib/api";
import { getSelectedOrg } from "@/lib/session";
import type { VPSInstance } from "@/lib/types";

type Params = { params: { id: string; action: string } };

const ALLOWED = new Set(["start", "stop", "console"]);

export async function POST(_request: Request, { params }: Params) {
  if (!ALLOWED.has(params.action)) {
    return NextResponse.json(
      { error: { message: "Acción no soportada" } },
      { status: 400 },
    );
  }
  const org = getSelectedOrg();
  if (!org) {
    return NextResponse.json(
      { error: { message: "Selecciona una organización primero" } },
      { status: 400 },
    );
  }
  const res = await routeApiCall<VPSInstance>({
    method: "POST",
    path: `/api/v1/vps/${params.id}/${params.action}`,
    tenantId: org,
  });
  return NextResponse.json(
    { data: res.data, error: res.error },
    { status: res.error ? res.status : params.action === "console" ? 201 : 200 },
  );
}
