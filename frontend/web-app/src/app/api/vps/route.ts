import { NextResponse } from "next/server";
import { routeApiCall } from "@/lib/api";
import { getSelectedOrg } from "@/lib/session";
import type { VPSInstance } from "@/lib/types";

function requireOrg() {
  const org = getSelectedOrg();
  if (!org) {
    return {
      response: NextResponse.json(
        { error: { message: "Selecciona una organización primero" } },
        { status: 400 },
      ),
      org: null,
    };
  }
  return { response: null, org };
}

export async function GET() {
  const { response, org } = requireOrg();
  if (response) return response;

  const res = await routeApiCall<{ vps_instances: VPSInstance[] }>({
    path: "/api/v1/vps",
    tenantId: org!,
  });
  return NextResponse.json(
    { data: res.data?.vps_instances ?? [], error: res.error },
    { status: res.error ? res.status : 200 },
  );
}
