import { create } from 'zustand';
import {
  applyHtmlTheme,
  loadThemeSetting,
  loadUIStyle,
  type ThemeSetting,
  type UIStyle,
} from './theme';

const THEME_KEY = 'theme';
const STYLE_KEY = 'ui-style';

type ThemeState = {
  theme: ThemeSetting;
  uiStyle: UIStyle;
  setTheme: (theme: ThemeSetting) => void;
  setUIStyle: (style: UIStyle) => void;
  rehydrateFromStorage: () => void;
};

function persistTheme(theme: ThemeSetting, uiStyle: UIStyle) {
  localStorage.setItem(THEME_KEY, theme);
  localStorage.setItem(STYLE_KEY, uiStyle);
  applyHtmlTheme(theme, uiStyle);
}

export const useThemeStore = create<ThemeState>((set, get) => ({
  theme: loadThemeSetting(),
  uiStyle: loadUIStyle(),

  setTheme: (theme) => {
    const { uiStyle } = get();
    persistTheme(theme, uiStyle);
    set({ theme });
  },

  setUIStyle: (uiStyle) => {
    const { theme } = get();
    persistTheme(theme, uiStyle);
    set({ uiStyle });
  },

  rehydrateFromStorage: () => {
    const theme = loadThemeSetting();
    const uiStyle = loadUIStyle();
    applyHtmlTheme(theme, uiStyle);
    set({ theme, uiStyle });
  },
}));

export const THEME_STORAGE_KEYS = [THEME_KEY, STYLE_KEY] as const;

/** Call once at app bootstrap so OS theme changes re-apply when theme === system. */
export function startThemeSystemWatch(): () => void {
  const mql = window.matchMedia?.('(prefers-color-scheme: dark)');
  if (!mql) return () => {};
  const handler = () => {
    const { theme, uiStyle } = useThemeStore.getState();
    if (theme === 'system') {
      applyHtmlTheme('system', uiStyle);
    }
  };
  mql.addEventListener('change', handler);
  return () => mql.removeEventListener('change', handler);
}
