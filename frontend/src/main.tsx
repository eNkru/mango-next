import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import { App } from './App';
import { routerBasename } from './lib/bootContext';
import { startPrefsStorageSync } from './lib/prefsSync';
import { applyHtmlTheme } from './lib/theme';
import { startThemeSystemWatch, useThemeStore } from './lib/themeStore';
// Side-effect: apply document lang from i18n store on import.
import './lib/i18n';
import './styles/fonts.css';
import './styles/tokens.css';
import './styles/shell.css';

const rootEl = document.getElementById('root');
if (!rootEl) {
  throw new Error('#root missing');
}

// Apply stored theme immediately; keep OS-system watch + cross-tab sync alive.
{
  const { theme, uiStyle } = useThemeStore.getState();
  applyHtmlTheme(theme, uiStyle);
  startThemeSystemWatch();
  startPrefsStorageSync();
}

createRoot(rootEl).render(
  <StrictMode>
    <BrowserRouter basename={routerBasename()}>
      <App />
    </BrowserRouter>
  </StrictMode>,
);
