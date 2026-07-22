import { createContext, useContext, useMemo, type ReactNode } from 'react';
import { normalizeBaseUrl, readBoot } from './boot';

export type BootSession = {
  baseUrl: string;
  isAdmin: boolean;
  version: string;
  pageName: string;
};

const BootContext = createContext<BootSession | null>(null);

export function BootProvider({ children }: { children: ReactNode }) {
  const session = useMemo((): BootSession => {
    const boot = readBoot();
    return {
      baseUrl: normalizeBaseUrl(boot.baseUrl),
      isAdmin: boot.isAdmin,
      version: boot.version,
      pageName: boot.pageName,
    };
  }, []);

  return <BootContext.Provider value={session}>{children}</BootContext.Provider>;
}

export function useBoot(): BootSession {
  const ctx = useContext(BootContext);
  if (!ctx) {
    // Fallback for tests / stray usage outside provider.
    const boot = readBoot();
    return {
      baseUrl: normalizeBaseUrl(boot.baseUrl),
      isAdmin: boot.isAdmin,
      version: boot.version,
      pageName: boot.pageName,
    };
  }
  return ctx;
}

/** BrowserRouter basename: '' for root, else without trailing slash. */
export function routerBasename(baseUrl?: string): string {
  const base = normalizeBaseUrl(baseUrl ?? readBoot().baseUrl);
  if (base === '/') return '';
  return base.replace(/\/$/, '');
}

/** Absolute app path for react-router navigate (starts with / or basename-relative path). */
export function appPath(path = ''): string {
  const rel = path.replace(/^\//, '');
  return rel ? `/${rel}` : '/';
}
