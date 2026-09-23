export const GATEWAY_URL =
  process.env.BOSSCLOUD_GATEWAY_URL?.replace(/\/$/, "") || "http://127.0.0.1:8080";

export const ACCESS_COOKIE = "bc_access";
export const REFRESH_COOKIE = "bc_refresh";
export const ORG_COOKIE = "bc_org";

export type GatewayResult<T> = {
  status: number;
  data: T | null;
  error: { code?: string; message?: string } | null;
};

type ForwardOptions = {
  method?: string;
  path: string;
  token?: string;
  tenantId?: string;
  body?: unknown;
  idempotencyKey?: string;
};

export async function forward<T = unknown>(
  opts: ForwardOptions,
): Promise<GatewayResult<T>> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };
  if (opts.token) headers["Authorization"] = `Bearer ${opts.token}`;
  if (opts.tenantId) headers["X-Tenant-ID"] = opts.tenantId;
  if (opts.idempotencyKey) headers["Idempotency-Key"] = opts.idempotencyKey;

  let res: Response;
  try {
    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), 15000);
    res = await fetch(`${GATEWAY_URL}${opts.path}`, {
      method: opts.method ?? "GET",
      headers,
      body: opts.body !== undefined ? JSON.stringify(opts.body) : undefined,
      cache: "no-store",
      signal: controller.signal,
    });
    clearTimeout(timer);
  } catch (err) {
    const aborted = err instanceof Error && err.name === "AbortError";
    return {
      status: aborted ? 504 : 502,
      data: null,
      error: {
        code: aborted ? "gateway_timeout" : "gateway_unreachable",
        message: aborted
          ? "El gateway tardó demasiado en responder. ¿Están todos los servicios en marcha?"
          : "No se pudo contactar con el gateway. ¿Están los servicios en marcha?",
      },
    };
  }

  const text = await res.text();
  let parsed: unknown = null;
  if (text) {
    try {
      parsed = JSON.parse(text);
    } catch {
      parsed = { message: text };
    }
  }

  if (res.ok) {
    return { status: res.status, data: (parsed as T) ?? null, error: null };
  }

  const errObj = (parsed as { code?: string; message?: string }) ?? {};
  return {
    status: res.status,
    data: null,
    error: { code: errObj.code, message: errObj.message ?? "Error inesperado" },
  };
}
