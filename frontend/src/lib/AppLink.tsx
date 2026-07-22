import type { ReactNode } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { appPath } from './bootContext';
import { baseUrl } from './baseUrl';

type AppLinkProps = {
  to: string;
  className?: string;
  children: ReactNode;
  replace?: boolean;
  'aria-label'?: string;
  title?: string;
};

/** SPA link for in-app routes. `to` is a path relative to BaseURL (e.g. "library", "book/x"). */
export function AppLink({ to, className, children, replace, ...rest }: AppLinkProps) {
  // Support query strings: "admin/user/edit?username=x"
  const [pathname, search = ''] = to.split('?');
  const dest = search ? `${appPath(pathname)}?${search}` : appPath(pathname);
  return (
    <Link to={dest} className={className} replace={replace} {...rest}>
      {children}
    </Link>
  );
}

/** Programmatic SPA navigation using the same path convention as baseUrl(). */
export function useAppNavigate() {
  const navigate = useNavigate();
  return (to: string, opts?: { replace?: boolean }) => {
    const [pathname, search = ''] = to.split('?');
    const dest = search ? `${appPath(pathname)}?${search}` : appPath(pathname);
    navigate(dest, { replace: opts?.replace });
  };
}

/** Use for download/logout/external — full document navigation. */
export function hardHref(path: string): string {
  return baseUrl(path);
}

export function isModifiedClick(event: MouseEvent): boolean {
  return event.metaKey || event.altKey || event.ctrlKey || event.shiftKey || event.button !== 0;
}
