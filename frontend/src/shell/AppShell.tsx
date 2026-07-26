import { useEffect, useRef, useState, type KeyboardEvent, type MutableRefObject, type ReactNode } from 'react';
import { AppLink } from '../lib/AppLink';
import { baseUrl } from '../lib/baseUrl';
import { useBoot } from '../lib/bootContext';
import { useI18n } from '../lib/i18n';
import { useThemeStore } from '../lib/themeStore';
import type { ThemeSetting, UIStyle } from '../lib/theme';
import { AlertHost } from './AlertHost';
import { GithubOctocat } from './GithubOctocat';
import { Icon } from './Icon';
import { icons } from './icons';

type MenuName = 'theme' | 'uiStyle' | 'language';

type AppShellProps = {
  title: string;
  subtitle?: string;
  children: ReactNode;
};

export function AppShell({ title, subtitle, children }: AppShellProps) {
  const { language, setLanguage, t } = useI18n();
  const boot = useBoot();
  const theme = useThemeStore((s) => s.theme);
  const uiStyle = useThemeStore((s) => s.uiStyle);
  const setTheme = useThemeStore((s) => s.setTheme);
  const setUIStyle = useThemeStore((s) => s.setUIStyle);
  const [openMenu, setOpenMenu] = useState<MenuName | null>(null);
  const toolsRef = useRef<HTMLDivElement>(null);
  const triggerRefs = useRef<Record<MenuName, HTMLButtonElement | null>>({
    theme: null,
    uiStyle: null,
    language: null,
  });
  const menuRefs = useRef<Record<MenuName, HTMLDivElement | null>>({
    theme: null,
    uiStyle: null,
    language: null,
  });

  useEffect(() => {
    document.title = `Mango - ${title}`;
  }, [language, title]);

  useEffect(() => {
    if (!openMenu) return;

    const dismissOnOutsidePointer = (event: PointerEvent) => {
      if (!toolsRef.current?.contains(event.target as Node)) {
        const trigger = triggerRefs.current[openMenu];
        setOpenMenu(null);
        trigger?.focus();
      }
    };
    document.addEventListener('pointerdown', dismissOnOutsidePointer);
    return () => document.removeEventListener('pointerdown', dismissOnOutsidePointer);
  }, [openMenu]);

  useEffect(() => {
    if (!openMenu) return;

    requestAnimationFrame(() => {
      menuRefs.current[openMenu]
        ?.querySelector<HTMLButtonElement>('[role="menuitemradio"][aria-checked="true"]')
        ?.focus();
    });
  }, [openMenu]);

  const closeMenu = (restoreFocus = true) => {
    if (openMenu && restoreFocus) triggerRefs.current[openMenu]?.focus();
    setOpenMenu(null);
  };

  const toggleMenu = (menu: MenuName) => {
    setOpenMenu((current) => (current === menu ? null : menu));
  };

  const handleMenuKeyDown = (event: KeyboardEvent<HTMLDivElement>, menu: MenuName) => {
    const items = Array.from(
      menuRefs.current[menu]?.querySelectorAll<HTMLButtonElement>('[role="menuitemradio"]') ?? [],
    );
    const currentIndex = items.indexOf(document.activeElement as HTMLButtonElement);
    if (event.key === 'Escape') {
      event.preventDefault();
      closeMenu();
    } else if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
      event.preventDefault();
      const offset = event.key === 'ArrowDown' ? 1 : -1;
      items[(currentIndex + offset + items.length) % items.length]?.focus();
    } else if (event.key === 'Home' || event.key === 'End') {
      event.preventDefault();
      items[event.key === 'Home' ? 0 : items.length - 1]?.focus();
    }
  };

  const openWithKeyboard = (event: KeyboardEvent<HTMLButtonElement>, menu: MenuName) => {
    if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
      event.preventDefault();
      setOpenMenu(menu);
      requestAnimationFrame(() => {
        const items = menuRefs.current[menu]?.querySelectorAll<HTMLButtonElement>('[role="menuitemradio"]');
        items?.[event.key === 'ArrowDown' ? 0 : items.length - 1]?.focus();
      });
    }
    if (event.key === 'Escape' && openMenu === menu) {
      event.preventDefault();
      closeMenu(false);
    }
  };

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
        <div className="mango-topbar__tools" ref={toolsRef}>
          <a
            className="mango-btn mango-btn--icon mango-topbar__external-link"
            href="https://github.com/eNkru/mango-next"
            target="_blank"
            rel="noopener noreferrer"
            aria-label={t('githubRepository')}
            title={t('githubRepository')}
            data-tooltip={t('githubRepository')}
            onClick={() => closeMenu(false)}
          >
            <GithubOctocat />
          </a>
          <a
            className="mango-btn mango-btn--icon mango-topbar__external-link"
            href="https://github.com/eNkru/mango-next/issues"
            target="_blank"
            rel="noopener noreferrer"
            aria-label={t('reportIssue')}
            title={t('reportIssue')}
            data-tooltip={t('reportIssue')}
            onClick={() => closeMenu(false)}
          >
            <Icon icon={icons.reportIssue} />
          </a>
          <PreferenceMenu
            name="theme"
            label={t('theme')}
            icon={icons.theme}
            openMenu={openMenu}
            toggleMenu={toggleMenu}
            triggerRefs={triggerRefs}
            menuRefs={menuRefs}
            onTriggerKeyDown={openWithKeyboard}
            onMenuKeyDown={handleMenuKeyDown}
            onSelect={(value) => setTheme(value as ThemeSetting)}
            value={theme}
            options={[['system', t('themeSystem')], ['light', t('themeLight')], ['dark', t('themeDark')]]}
            closeMenu={closeMenu}
          />
          <PreferenceMenu
            name="uiStyle"
            label={t('uiStyle')}
            icon={icons.uiStyle}
            openMenu={openMenu}
            toggleMenu={toggleMenu}
            triggerRefs={triggerRefs}
            menuRefs={menuRefs}
            onTriggerKeyDown={openWithKeyboard}
            onMenuKeyDown={handleMenuKeyDown}
            onSelect={(value) => setUIStyle(value as UIStyle)}
            value={uiStyle}
            options={[['comic', t('uiStyleComic')], ['flat', t('uiStyleFlat')]]}
            closeMenu={closeMenu}
          />
          <PreferenceMenu
            name="language"
            label={t('language')}
            icon={icons.language}
            openMenu={openMenu}
            toggleMenu={toggleMenu}
            triggerRefs={triggerRefs}
            menuRefs={menuRefs}
            onTriggerKeyDown={openWithKeyboard}
            onMenuKeyDown={handleMenuKeyDown}
            onSelect={(value) => setLanguage(value as typeof language)}
            value={language}
            options={[
              ['zh-cn', t('languageSimplifiedChinese')],
              ['zh-tw', t('languageTraditionalChinese')],
              ['en', t('languageEnglish')],
            ]}
            closeMenu={closeMenu}
          />
          <a className="mango-btn mango-btn--icon" href={baseUrl('logout')} aria-label={t('logout')}>
            <Icon icon={icons.logout} size={16} />
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

type PreferenceMenuProps = {
  name: MenuName;
  label: string;
  icon: typeof icons.theme;
  openMenu: MenuName | null;
  toggleMenu: (menu: MenuName) => void;
  triggerRefs: MutableRefObject<Record<MenuName, HTMLButtonElement | null>>;
  menuRefs: MutableRefObject<Record<MenuName, HTMLDivElement | null>>;
  onTriggerKeyDown: (event: KeyboardEvent<HTMLButtonElement>, menu: MenuName) => void;
  onMenuKeyDown: (event: KeyboardEvent<HTMLDivElement>, menu: MenuName) => void;
  onSelect: (value: string) => void;
  value: string;
  options: [string, string][];
  closeMenu: () => void;
};

function PreferenceMenu({
  name,
  label,
  icon,
  openMenu,
  toggleMenu,
  triggerRefs,
  menuRefs,
  onTriggerKeyDown,
  onMenuKeyDown,
  onSelect,
  value,
  options,
  closeMenu,
}: PreferenceMenuProps) {
  const isOpen = openMenu === name;
  const menuId = `mango-${name}-menu`;
  return (
    <div className="mango-topbar__menu-wrap">
      <button
        ref={(element) => { triggerRefs.current[name] = element; }}
        className="mango-btn mango-btn--icon"
        type="button"
        aria-label={label}
        title={label}
        aria-haspopup="menu"
        aria-expanded={isOpen}
        aria-controls={menuId}
        onClick={() => toggleMenu(name)}
        onKeyDown={(event) => onTriggerKeyDown(event, name)}
      >
        <Icon icon={icon} />
      </button>
      {isOpen ? (
        <div
          ref={(element) => { menuRefs.current[name] = element; }}
          id={menuId}
          className="mango-topbar__menu"
          role="menu"
          aria-label={label}
          onKeyDown={(event) => onMenuKeyDown(event, name)}
        >
          {options.map(([optionValue, optionLabel]) => (
            <button
              key={optionValue}
              type="button"
              role="menuitemradio"
              aria-checked={value === optionValue}
              className="mango-topbar__menu-item"
              onClick={() => {
                onSelect(optionValue);
                closeMenu();
              }}
            >
              {optionLabel}
            </button>
          ))}
        </div>
      ) : null}
    </div>
  );
}
