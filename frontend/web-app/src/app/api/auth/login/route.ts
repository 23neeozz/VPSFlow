import { NextResponse } from "next/server";
import { forward, ACCESS_COOKIE, REFRESH_COOKIE } from "@/lib/gateway";
import type { AuthTokens } from "@/lib/types";

const COOKIE_BASE = {
  httpOnly: true,
  sameSite: "lax" as const,
  secure: process.env.NODE_ENV === "production",
  path: "/",
};

function authOkResponse(tokens: AuthTokens) {
  const res = NextResponse.json({ ok: true });
  res.cookies.set(ACCESS_COOKIE, tokens.access_token, {
    ...COOKIE_BASE,
    maxAge: 60 * 20,
  });
  res.cookies.set(REFRESH_COOKIE, tokens.refresh_token, {
    ...COOKIE_BASE,
    maxAge: 60 * 60 * 24 * 30,
  });
  return res;
}

export async function POST(request: Request) {
  const body = await request.json().catch(() => ({}));
  const res = await forward<AuthTokens>({
    method: "POST",
    path: "/api/v1/auth/login",
    body,
  });

  if (res.data?.access_token) {
    return authOkResponse(res.data);
  }

  return NextResponse.json(
    { ok: false, error: res.error },
    { status: res.status || 400 },
  );
}
