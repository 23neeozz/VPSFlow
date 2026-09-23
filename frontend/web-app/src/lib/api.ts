import { forward, type GatewayResult } from "@/lib/gateway";
import {
  getAccessToken,
  getRefreshToken,
  setAuthCookies,
  clearAuthCookies,
} from "@/lib/session";
import type { AuthTokens } from "@/lib/types";

type CallOptions = {
  method?: string;
  path: string;
  tenantId?: string;
  body?: unknown;
  idempotencyKey?: string;
  /** Only route handlers and server actions may persist refreshed tokens. */
  mutateCookies?: boolean;
};

async function tryRefresh(mutateCookies: boolean): Promise<string | null> {
  const refreshToken = getRefreshToken();
  if (!refreshToken) return null;

  const res = await forward<AuthTokens>({
    method: "POST",
    path: "/api/v1/auth/refresh",
    body: { refresh_token: refreshToken },
  });

  if (res.data?.access_token) {
    if (mutateCookies) {
      setAuthCookies(res.data.access_token, res.data.refresh_token);
    }
    return res.data.access_token;
  }
  if (mutateCookies) {
    clearAuthCookies();
  }
  return null;
}

/**
 * Authenticated call to the gateway. Transparently retries once with a
 * refreshed access token when the current one is expired/unauthorized.
 */
export async function apiCall<T = unknown>(
  opts: CallOptions,
): Promise<GatewayResult<T>> {
  const mutateCookies = opts.mutateCookies ?? false;
  let token = getAccessToken();

  let res = await forward<T>({ ...opts, token });

  if (res.status === 401) {
    const refreshed = await tryRefresh(mutateCookies);
    if (refreshed) {
      token = refreshed;
      res = await forward<T>({ ...opts, token });
    }
  }

  return res;
}

/** Use from route handlers so refreshed tokens are persisted in cookies. */
export async function routeApiCall<T = unknown>(
  opts: Omit<CallOptions, "mutateCookies">,
): Promise<GatewayResult<T>> {
  return apiCall<T>({ ...opts, mutateCookies: true });
}
