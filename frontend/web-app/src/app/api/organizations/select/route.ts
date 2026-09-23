import { NextResponse } from "next/server";
import { setSelectedOrg } from "@/lib/session";

export async function POST(request: Request) {
  const body = await request.json().catch(() => ({}));
  const orgId = typeof body.org_id === "string" ? body.org_id : "";
  if (!orgId) {
    return NextResponse.json(
      { ok: false, error: { message: "org_id requerido" } },
      { status: 400 },
    );
  }
  setSelectedOrg(orgId);
  return NextResponse.json({ ok: true });
}
