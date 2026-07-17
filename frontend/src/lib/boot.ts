export type MangoBoot = {
  baseUrl: string;
  pageId: string;
  pageName: string;
  isAdmin: boolean;
  version: string;
  /** Optional page-specific payload, e.g. edit-target username. */
  username?: string;
};

const DEFAULT_BOOT: MangoBoot = {
  baseUrl: '/',
  pageId: 'react-preview',
  pageName: 'react-preview',
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
