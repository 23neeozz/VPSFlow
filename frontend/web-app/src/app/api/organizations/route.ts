import { NextResponse } from "next/server";
import { routeApiCall } from "@/lib/api";
import type { Organization } from "@/lib/types";

export async function GET() {
  const res = await routeApiCall<{ data: Organization[] }>({
    path: "/api/v1/organizations",
  });
  return NextResponse.json(
    { data: res.data?.data ?? [], error: res.error },
    { status: res.error ? res.status : 200 },
  );
}

export async function POST(request: Request) {
  const body = await request.json().catch(() => ({}));
  const res = await routeApiCall<Organization>({
    method: "POST",
    path: "/api/v1/organizations",
    body,
  });
  return NextResponse.json(
    { data: res.data, error: res.error },
    { status: res.error ? res.status : 201 },
  );
}
