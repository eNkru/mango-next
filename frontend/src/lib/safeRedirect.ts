export function safeRedirectPath(callback: string | null | undefined): string {
  const raw = (callback ?? '').trim();
  if (!raw) return '/';
  if (!raw.startsWith('/') || raw.startsWith('//')) return '/';
  if (raw.includes('\\')) return '/';
  try {
    const u = new URL(raw, 'http://mango.local');
    if (u.origin !== 'http://mango.local') return '/';
    if (!u.pathname.startsWith('/') || u.pathname.startsWith('//')) return '/';
    let out = u.pathname;
    if (u.search) out += u.search;
    if (u.hash) out += u.hash;
    return out || '/';
  } catch {
    return '/';
  }
}

export function resolvePostLoginHref(baseUrl: string, callback: string | null | undefined): string {
  const safe = safeRedirectPath(callback);
  const base = normalizeJoinBase(baseUrl);
  if (safe === '/') return base;
  const rel = safe.replace(/^\//, '');
  return `${base}${rel}`;
}

function normalizeJoinBase(base: string): string {
  if (!base) return '/';
  let b = base.startsWith('/') ? base : `/${base}`;
  if (!b.endsWith('/')) b = `${b}/`;
  return b;
}
