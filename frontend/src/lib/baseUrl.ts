import { normalizeBaseUrl, readBoot } from './boot';

/** Join BaseURL with a path segment that may or may not start with `/`. */
export function baseUrl(path = ''): string {
  const base = normalizeBaseUrl(readBoot().baseUrl);
  const rel = path.replace(/^\//, '');
  if (!rel) return base;
  return `${base}${rel}`;
}
