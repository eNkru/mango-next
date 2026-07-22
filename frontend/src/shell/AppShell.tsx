import { useEffect, useState, type ReactNode } from 'react';
import { AppLink } from '../lib/AppLink';
import { baseUrl } from '../lib/baseUrl';
import { useBoot } from '../lib/bootContext';
import { useI18n } from '../lib/i18n';
import {
  applyHtmlTheme,
  loadThemeSetting,
  loadUIStyle,
  saveThemeSetting,
  saveUIStyle,
  watchSystemTheme,
  type ThemeSetting,
  type UIStyle,
} from '../lib/theme';
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
  const [theme, setTheme] = useState<ThemeSetting>(loadThemeSetting);
  const [uiStyle, setUiStyle] = useState<UIStyle>(loadUIStyle);

  useEffect(() => {
    document.title = `Mango - ${title}`;
  }, [language, title]);

  useEffect(() => {
    applyHtmlTheme(theme, uiStyle);
  }, [theme, uiStyle]);

  useEffect(() => watchSystemTheme(), []);

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
                const next = event.target.value as ThemeSetting;
                setTheme(next);
                saveThemeSetting(next);
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
                const next = event.target.value as UIStyle;
                setUiStyle(next);
                saveUIStyle(next);
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
