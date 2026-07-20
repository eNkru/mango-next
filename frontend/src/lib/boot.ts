export type MangoBoot = {
  baseUrl: string;
  pageId: string;
  pageName: string;
  isAdmin: boolean;
  version: string;
  /** Optional page-specific payload, e.g. edit-target username. */
  username?: string;
  /** Optional tag name for tag detail pages. */
  tag?: string;
  showHidden?: boolean;
  /** Optional post-login redirect path (same-app relative). */
  callback?: string;
  /** Title id for the React title-detail route. */
  titleId?: string;
  /** Reader route identity (title / entry / one-based page). */
  tid?: string;
  eid?: string;
  page?: number;
};

const DEFAULT_BOOT: MangoBoot = {
  baseUrl: '/',
  pageId: 'home',
  pageName: 'home',
  isAdmin: true,
  version: 'dev',
};

export function readBoot(): MangoBoot {
  const el = document.getElementById('mango-boot');
  if (!el?.textContent) {
    return DEFAULT_BOOT;
  }
  try {
    const parsed = JSON.parse(el.textContent) as Partial<MangoBoot>;
    return {
      baseUrl: normalizeBaseUrl(parsed.baseUrl ?? '/'),
      pageId: parsed.pageId ?? DEFAULT_BOOT.pageId,
      pageName: parsed.pageName ?? DEFAULT_BOOT.pageName,
      isAdmin: Boolean(parsed.isAdmin),
      version: parsed.version ?? DEFAULT_BOOT.version,
      username: typeof parsed.username === 'string' ? parsed.username : undefined,
      tag: typeof parsed.tag === 'string' ? parsed.tag : undefined,
      showHidden: Boolean(parsed.showHidden),
      callback: typeof parsed.callback === 'string' ? parsed.callback : undefined,
      titleId: typeof parsed.titleId === 'string' ? parsed.titleId : undefined,
      tid: typeof parsed.tid === 'string' ? parsed.tid : undefined,
      eid: typeof parsed.eid === 'string' ? parsed.eid : undefined,
      page: typeof parsed.page === 'number' && Number.isFinite(parsed.page) ? parsed.page : undefined,
    };
  } catch {
    return DEFAULT_BOOT;
  }
}

export function normalizeBaseUrl(base: string): string {
  if (!base) return '/';
  if (!base.startsWith('/')) base = `/${base}`;
  if (!base.endsWith('/')) base = `${base}/`;
  return base;
}
