import { baseUrl } from './baseUrl';

export class ApiError extends Error {
  status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

type JsonEnvelope = {
  success?: boolean;
  error?: string;
  data?: unknown;
  [key: string]: unknown;
};

export async function apiFetch<T = JsonEnvelope>(
  path: string,
  init: RequestInit = {},
): Promise<T> {
  const url = baseUrl(path);
  const headers = new Headers(init.headers);
  if (!headers.has('Accept')) {
    headers.set('Accept', 'application/json');
  }
  if (init.body && !(init.body instanceof FormData) && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }

  const res = await fetch(url, {
    credentials: 'same-origin',
    ...init,
    headers,
  });

  let payload: JsonEnvelope | null = null;
  const text = await res.text();
  if (text) {
    try {
      payload = JSON.parse(text) as JsonEnvelope;
    } catch {
      payload = null;
    }
  }

  if (!res.ok) {
    const msg =
      (payload && typeof payload.error === 'string' && payload.error) ||
      `Request failed: ${res.status} ${res.statusText}`;
    throw new ApiError(msg, res.status);
  }

  if (payload && payload.success === false) {
    throw new ApiError(payload.error || 'Request failed', res.status);
  }

  return (payload ?? {}) as T;
}
