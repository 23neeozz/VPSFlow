import { cookies } from "next/headers";
import {
  ACCESS_COOKIE,
  ORG_COOKIE,
  REFRESH_COOKIE,
} from "@/lib/gateway";

const COOKIE_BASE = {
  httpOnly: true,
  sameSite: "lax" as const,
  secure: process.env.NODE_ENV === "production",
  path: "/",
};

export function setAuthCookies(accessToken: string, refreshToken: string) {
  const store = cookies();
  store.set(ACCESS_COOKIE, accessToken, { ...COOKIE_BASE, maxAge: 60 * 20 });
  store.set(REFRESH_COOKIE, refreshToken, {
    ...COOKIE_BASE,
    maxAge: 60 * 60 * 24 * 30,
  });
}

export function clearAuthCookies() {
  const store = cookies();
  store.delete(ACCESS_COOKIE);
  store.delete(REFRESH_COOKIE);
  store.delete(ORG_COOKIE);
}

export function getAccessToken(): string | undefined {
  return cookies().get(ACCESS_COOKIE)?.value;
}

export function getRefreshToken(): string | undefined {
  return cookies().get(REFRESH_COOKIE)?.value;
}

export function getSelectedOrg(): string | undefined {
  return cookies().get(ORG_COOKIE)?.value;
}

export function setSelectedOrg(orgId: string) {
  cookies().set(ORG_COOKIE, orgId, {
    ...COOKIE_BASE,
    httpOnly: false,
    maxAge: 60 * 60 * 24 * 30,
  });
}

export function isAuthenticated(): boolean {
  return Boolean(getAccessToken());
}
