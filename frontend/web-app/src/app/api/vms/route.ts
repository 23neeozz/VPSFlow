import { NextResponse } from "next/server";
import { routeApiCall } from "@/lib/api";
import { getSelectedOrg } from "@/lib/session";
import type { VirtualMachine } from "@/lib/types";

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

  const res = await routeApiCall<{ virtual_machines: VirtualMachine[] }>({
    path: "/api/v1/virtual-machines",
    tenantId: org!,
  });
  return NextResponse.json(
    { data: res.data?.virtual_machines ?? [], error: res.error },
    { status: res.error ? res.status : 200 },
  );
}

export async function POST(request: Request) {
  const { response, org } = requireOrg();
  if (response) return response;

  const body = await request.json().catch(() => ({}));
  const res = await routeApiCall<VirtualMachine>({
    method: "POST",
    path: "/api/v1/virtual-machines",
    tenantId: org!,
    body,
    idempotencyKey:
      typeof crypto !== "undefined" && "randomUUID" in crypto
        ? crypto.randomUUID()
        : undefined,
  });
  return NextResponse.json(
    { data: res.data, error: res.error },
    { status: res.error ? res.status : 202 },
  );
}
