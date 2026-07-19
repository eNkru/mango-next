export type ThemeSetting = 'dark' | 'light' | 'system';
export type UIStyle = 'comic' | 'flat';

const THEME_KEY = 'theme';
const STYLE_KEY = 'ui-style';

export function loadThemeSetting(): ThemeSetting {
  const v = localStorage.getItem(THEME_KEY);
  if (v === 'dark' || v === 'light' || v === 'system') return v;
  return 'system';
}

export function loadUIStyle(): UIStyle {
  const v = localStorage.getItem(STYLE_KEY);
  if (v === 'flat' || v === 'comic') return v;
  return 'comic';
}

export function systemPrefersDark(): boolean {
  return Boolean(window.matchMedia?.('(prefers-color-scheme: dark)').matches);
}

export function resolveDark(theme: ThemeSetting): boolean {
  return theme === 'dark' || (theme === 'system' && systemPrefersDark());
}

/** Apply comic/flat + dark markers on <html>, matching react-shell.tmpl FOUC. */
export function applyHtmlTheme(theme: ThemeSetting = loadThemeSetting(), style: UIStyle = loadUIStyle()): void {
  const root = document.documentElement;
  const dark = resolveDark(theme);
  root.classList.remove('comic-theme', 'comic-theme-dark', 'flat-theme', 'flat-theme-dark');
  if (style === 'flat') {
    root.classList.add('flat-theme');
    if (dark) root.classList.add('flat-theme-dark');
    root.style.background = dark ? '#141414' : '';
  } else {
    root.classList.add('comic-theme');
    if (dark) root.classList.add('comic-theme-dark');
    root.style.background = '';
  }
}

export function saveThemeSetting(theme: ThemeSetting): void {
  localStorage.setItem(THEME_KEY, theme);
  applyHtmlTheme(theme, loadUIStyle());
}

export function saveUIStyle(style: UIStyle): void {
  localStorage.setItem(STYLE_KEY, style);
  applyHtmlTheme(loadThemeSetting(), style);
}

/** Re-apply when OS theme changes and user chose "system". */
export function watchSystemTheme(onChange?: () => void): () => void {
  const mql = window.matchMedia?.('(prefers-color-scheme: dark)');
  if (!mql) return () => {};
  const handler = () => {
    if (loadThemeSetting() === 'system') {
      applyHtmlTheme('system', loadUIStyle());
      onChange?.();
    }
  };
  mql.addEventListener('change', handler);
  return () => mql.removeEventListener('change', handler);
}
