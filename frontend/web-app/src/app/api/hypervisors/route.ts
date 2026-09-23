import { NextResponse } from "next/server";
import { routeApiCall } from "@/lib/api";
import type { Hypervisor } from "@/lib/types";

export async function GET() {
  const res = await routeApiCall<{ hypervisors: Hypervisor[] }>({
    path: "/api/v1/hypervisors",
  });
  return NextResponse.json(
    { data: res.data?.hypervisors ?? [], error: res.error },
    { status: res.error ? res.status : 200 },
  );
}
