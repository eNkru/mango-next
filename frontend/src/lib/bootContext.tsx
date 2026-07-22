import { create } from 'zustand';
import { normalizeBaseUrl, readBoot } from './boot';

export type BootSession = {
  baseUrl: string;
  isAdmin: boolean;
  version: string;
  pageName: string;
};

function loadBootSession(): BootSession {
  const boot = readBoot();
  return {
    baseUrl: normalizeBaseUrl(boot.baseUrl),
    isAdmin: boot.isAdmin,
    version: boot.version,
    pageName: boot.pageName,
  };
}

type BootState = BootSession;

export const useBootStore = create<BootState>(() => loadBootSession());

export function useBoot(): BootSession {
  const baseUrl = useBootStore((s) => s.baseUrl);
  const isAdmin = useBootStore((s) => s.isAdmin);
  const version = useBootStore((s) => s.version);
  const pageName = useBootStore((s) => s.pageName);
  return { baseUrl, isAdmin, version, pageName };
}

/** BrowserRouter basename: '' for root, else without trailing slash. */
export function routerBasename(baseUrl?: string): string {
  const base = normalizeBaseUrl(baseUrl ?? useBootStore.getState().baseUrl);
  if (base === '/') return '';
  return base.replace(/\/$/, '');
}

/** Absolute app path for react-router navigate (starts with /). */
export function appPath(path = ''): string {
  const rel = path.replace(/^\//, '');
  return rel ? `/${rel}` : '/';
}
