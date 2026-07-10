// Central API client using fetch. Reads the base URL from env with a sensible
// default for local development (Vite proxy forwards /api to the backend).

const BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1';

/** Error thrown when the API returns a non-success envelope. */
export class ApiError extends Error {
  status: number;
  errors?: { field: string; reason: string }[];

  constructor(message: string, status: number, errors?: { field: string; reason: string }[]) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.errors = errors;
  }
}

interface RequestOptions {
  method?: string;
  body?: unknown;
  signal?: AbortSignal;
}

/**
 * request is the low-level fetch wrapper that handles JSON encoding/decoding
 * and the standard response envelope. It throws ApiError on non-success responses.
 */
export async function request<T>(path: string, opts: RequestOptions = {}): Promise<T> {
  const { method = 'GET', body, signal } = opts;

  const headers: Record<string, string> = {
    Accept: 'application/json',
  };
  if (body !== undefined) {
    headers['Content-Type'] = 'application/json';
  }

  const res = await fetch(`${BASE_URL}${path}`, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
    signal,
  });

  // Attempt to parse JSON regardless of status — our envelope is always JSON.
  let payload: unknown = null;
  const text = await res.text();
  if (text) {
    try {
      payload = JSON.parse(text);
    } catch {
      // Non-JSON response (e.g. 502 from a proxy) — fall through to generic error.
    }
  }

  if (!res.ok) {
    const errPayload = payload as { message?: string; errors?: { field: string; reason: string }[] } | null;
    throw new ApiError(
      errPayload?.message || `Request failed with status ${res.status}`,
      res.status,
      errPayload?.errors,
    );
  }

  // The envelope guarantees { success, message, data, meta? }.
  const envelope = payload as { success: boolean; message: string; data: T; meta?: unknown };
  return envelope.data;
}

/** requestFull returns the full envelope (data + meta) for paginated endpoints. */
export async function requestFull<T>(
  path: string,
  opts: RequestOptions = {},
): Promise<{ data: T; meta?: { page: number; limit: number; total: number } }> {
  const { method = 'GET', body, signal } = opts;

  const headers: Record<string, string> = { Accept: 'application/json' };
  if (body !== undefined) headers['Content-Type'] = 'application/json';

  const res = await fetch(`${BASE_URL}${path}`, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
    signal,
  });

  const text = await res.text();
  let payload: unknown = null;
  if (text) {
    try {
      payload = JSON.parse(text);
    } catch {
      /* ignore */
    }
  }

  if (!res.ok) {
    const errPayload = payload as { message?: string; errors?: { field: string; reason: string }[] } | null;
    throw new ApiError(
      errPayload?.message || `Request failed with status ${res.status}`,
      res.status,
      errPayload?.errors,
    );
  }

  const envelope = payload as {
    success: boolean;
    message: string;
    data: T;
    meta?: { page: number; limit: number; total: number };
  };
  return { data: envelope.data, meta: envelope.meta };
}
