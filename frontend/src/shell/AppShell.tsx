import type { ReactNode } from 'react';
import { baseUrl } from '../lib/baseUrl';
import { AlertHost } from './AlertHost';

type AppShellProps = {
  title: string;
  subtitle?: string;
  children: ReactNode;
};

export function AppShell({ title, subtitle, children }: AppShellProps) {
  return (
    <>
      <header className="mango-topbar" role="banner">
        <a className="mango-topbar__brand" href={baseUrl()}>
          <span>Mango</span>
        </a>
        <nav aria-label="主导航">
          <ul className="mango-topbar__nav">
            <li>
              <a href={baseUrl()}>主页</a>
            </li>
            <li>
              <a href={baseUrl('library')}>资料库</a>
            </li>
            <li>
              <a href={baseUrl('tags')}>标签</a>
            </li>
            <li>
              <a href={baseUrl('admin')}>管理员</a>
            </li>
          </ul>
        </nav>
      </header>
      <main className="mango-shell">
        <header className="mango-page-header">
          <h1>{title}</h1>
          {subtitle ? <p>{subtitle}</p> : null}
        </header>
        {children}
      </main>
      <AlertHost />
    </>
  );
}
