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

/**
 * Dev-only fallback when Go did not inject #mango-boot (Vite `npm run dev`).
 * Production always has mango-boot from react-shell; this path is unused there.
 */
export function bootFromPathname(pathname: string, search = ''): Partial<MangoBoot> {
  const path = pathname.replace(/\/+$/, '') || '/';
  const qs = new URLSearchParams(search.startsWith('?') ? search.slice(1) : search);

  if (path === '/' || path === '') {
    return { pageId: 'home', pageName: 'home' };
  }
  if (path === '/login') {
    const callback = qs.get('callback');
    return {
      pageId: 'login',
      pageName: 'login',
      isAdmin: false,
      ...(callback ? { callback } : {}),
    };
  }
  if (path === '/library') {
    return { pageId: 'library', pageName: 'library' };
  }
  if (path === '/tags') {
    return { pageId: 'tags-index', pageName: 'tags' };
  }
  {
    const m = path.match(/^\/tags\/([^/]+)$/);
    if (m) {
      return {
        pageId: 'tag-detail',
        pageName: 'tag',
        tag: decodeURIComponent(m[1]),
        showHidden: qs.get('show_hidden') === '1',
      };
    }
  }
  {
    const m = path.match(/^\/book\/([^/]+)$/);
    if (m) {
      return {
        pageId: 'title-detail',
        pageName: 'title',
        titleId: decodeURIComponent(m[1]),
      };
    }
  }
  {
    const m = path.match(/^\/reader\/([^/]+)\/([^/]+)(?:\/(\d+))?$/);
    if (m) {
      const page = m[3] ? Number(m[3]) : 1;
      return {
        pageId: 'reader',
        pageName: 'reader',
        tid: decodeURIComponent(m[1]),
        eid: decodeURIComponent(m[2]),
        page: Number.isFinite(page) && page >= 1 ? page : 1,
      };
    }
  }
  if (path === '/admin' || path === '/admin/') {
    return { pageId: 'admin', pageName: 'admin', isAdmin: true };
  }
  if (path === '/admin/user') {
    return { pageId: 'user-list', pageName: 'user-list', isAdmin: true };
  }
  if (path === '/admin/user/edit') {
    const username = qs.get('username');
    return {
      pageId: 'user-edit',
      pageName: 'user-edit',
      isAdmin: true,
      ...(username ? { username } : {}),
    };
  }
  if (path === '/admin/missing') {
    return { pageId: 'missing-items', pageName: 'missing-items', isAdmin: true };
  }

  return { pageId: 'home', pageName: 'home' };
}

function mergeBoot(base: MangoBoot, partial: Partial<MangoBoot>): MangoBoot {
  return {
    baseUrl: normalizeBaseUrl(partial.baseUrl ?? base.baseUrl),
    pageId: partial.pageId ?? base.pageId,
    pageName: partial.pageName ?? base.pageName,
    isAdmin: partial.isAdmin !== undefined ? Boolean(partial.isAdmin) : base.isAdmin,
    version: partial.version ?? base.version,
    username: typeof partial.username === 'string' ? partial.username : base.username,
    tag: typeof partial.tag === 'string' ? partial.tag : base.tag,
    showHidden: partial.showHidden !== undefined ? Boolean(partial.showHidden) : base.showHidden,
    callback: typeof partial.callback === 'string' ? partial.callback : base.callback,
    titleId: typeof partial.titleId === 'string' ? partial.titleId : base.titleId,
    tid: typeof partial.tid === 'string' ? partial.tid : base.tid,
    eid: typeof partial.eid === 'string' ? partial.eid : base.eid,
    page:
      typeof partial.page === 'number' && Number.isFinite(partial.page)
        ? partial.page
        : base.page,
  };
}

export function readBoot(): MangoBoot {
  const el = document.getElementById('mango-boot');
  if (!el?.textContent) {
    // Vite dev: no Go shell — infer page from URL (does not run when mango-boot exists).
    if (typeof window !== 'undefined') {
      return mergeBoot(DEFAULT_BOOT, bootFromPathname(window.location.pathname, window.location.search));
    }
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
