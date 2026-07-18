import { useEffect, type ReactNode } from 'react';
import { baseUrl } from '../lib/baseUrl';
import { readBoot } from '../lib/boot';
import { useI18n, type Language } from '../lib/i18n';
import { AlertHost } from './AlertHost';

type AppShellProps = {
  title: string;
  subtitle?: string;
  children: ReactNode;
};

export function AppShell({ title, subtitle, children }: AppShellProps) {
  const { language, setLanguage, t } = useI18n();
  const boot = readBoot();
  useEffect(() => { document.title = `Mango - ${title}`; }, [language, title]);
  return (
    <>
      <header className="mango-topbar" role="banner">
        <a className="mango-topbar__brand" href={baseUrl()}>
          <span>Mango</span>
        </a>
        <nav aria-label={t('home')}>
          <ul className="mango-topbar__nav">
            <li>
              <a href={baseUrl()}>{t('home')}</a>
            </li>
            <li>
              <a href={baseUrl('library')}>{t('library')}</a>
            </li>
            <li>
              <a href={baseUrl('tags')}>{t('tags')}</a>
            </li>
            {boot.isAdmin ? <li><a href={baseUrl('admin')}>{t('admin')}</a></li> : null}
          </ul>
        </nav>
        <div className="mango-topbar__tools">
          <label className="mango-language"><span className="sr-only">{t('language')}</span><select value={language} onChange={(event) => setLanguage(event.target.value as Language)} aria-label={t('language')}><option value="zh-cn">简体中文</option><option value="zh-tw">繁體中文</option><option value="en">English</option></select></label>
          <a className="mango-topbar__logout" href={baseUrl('logout')}>{t('logout')}</a>
        </div>
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
