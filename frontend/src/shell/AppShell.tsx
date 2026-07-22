import { useEffect, type ReactNode } from 'react';
import { AppLink } from '../lib/AppLink';
import { baseUrl } from '../lib/baseUrl';
import { useBoot } from '../lib/bootContext';
import { useI18n } from '../lib/i18n';
import { useThemeStore } from '../lib/themeStore';
import type { ThemeSetting, UIStyle } from '../lib/theme';
import { AlertHost } from './AlertHost';
import { Icon } from './Icon';
import { icons } from './icons';
import { LanguageSelect } from './LanguageSelect';

type AppShellProps = {
  title: string;
  subtitle?: string;
  children: ReactNode;
};

export function AppShell({ title, subtitle, children }: AppShellProps) {
  const { language, t } = useI18n();
  const boot = useBoot();
  const theme = useThemeStore((s) => s.theme);
  const uiStyle = useThemeStore((s) => s.uiStyle);
  const setTheme = useThemeStore((s) => s.setTheme);
  const setUIStyle = useThemeStore((s) => s.setUIStyle);

  useEffect(() => {
    document.title = `Mango - ${title}`;
  }, [language, title]);

  return (
    <>
      <header className="mango-topbar" role="banner">
        <AppLink className="mango-topbar__brand" to="">
          <img
            className="mango-topbar__mark"
            src={baseUrl('img/icons/mango-mark.svg')}
            alt=""
            width={28}
            height={28}
          />
          <span className="mango-topbar__wordmark">Mango</span>
        </AppLink>
        <nav aria-label={t('home')}>
          <ul className="mango-topbar__nav">
            <li>
              <AppLink to="">
                <Icon icon={icons.home} size={16} />
                {t('home')}
              </AppLink>
            </li>
            <li>
              <AppLink to="library">
                <Icon icon={icons.library} size={16} />
                {t('library')}
              </AppLink>
            </li>
            <li>
              <AppLink to="tags">
                <Icon icon={icons.tags} size={16} />
                {t('tags')}
              </AppLink>
            </li>
            {boot.isAdmin ? (
              <li>
                <AppLink to="admin">
                  <Icon icon={icons.admin} size={16} />
                  {t('admin')}
                </AppLink>
              </li>
            ) : null}
          </ul>
        </nav>
        <div className="mango-topbar__tools">
          <label className="mango-language">
            <span className="sr-only">{t('theme')}</span>
            <select
              value={theme}
              onChange={(event) => {
                setTheme(event.target.value as ThemeSetting);
              }}
              aria-label={t('theme')}
            >
              <option value="system">{t('themeSystem')}</option>
              <option value="light">{t('themeLight')}</option>
              <option value="dark">{t('themeDark')}</option>
            </select>
          </label>
          <label className="mango-language">
            <span className="sr-only">{t('uiStyle')}</span>
            <select
              value={uiStyle}
              onChange={(event) => {
                setUIStyle(event.target.value as UIStyle);
              }}
              aria-label={t('uiStyle')}
            >
              <option value="comic">{t('uiStyleComic')}</option>
              <option value="flat">{t('uiStyleFlat')}</option>
            </select>
          </label>
          <LanguageSelect />
          <a className="mango-topbar__logout" href={baseUrl('logout')} aria-label={t('logout')}>
            <Icon icon={icons.logout} size={16} />
            <span className="mango-topbar__logout-label">{t('logout')}</span>
          </a>
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
