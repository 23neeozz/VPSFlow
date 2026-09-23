import { NextResponse } from "next/server";
import { routeApiCall } from "@/lib/api";
import { clearAuthCookies } from "@/lib/session";

export async function POST() {
  await routeApiCall({ method: "POST", path: "/api/v1/auth/logout" });
  clearAuthCookies();
  return NextResponse.json({ ok: true });
}
