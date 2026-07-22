import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import { App } from './App';
import { BootProvider, routerBasename } from './lib/bootContext';
import { I18nProvider } from './lib/i18n';
import { startPrefsStorageSync } from './lib/prefsSync';
import { applyHtmlTheme } from './lib/theme';
import { startThemeSystemWatch, useThemeStore } from './lib/themeStore';
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
    <BootProvider>
      <BrowserRouter basename={routerBasename()}>
        <I18nProvider>
          <App />
        </I18nProvider>
      </BrowserRouter>
    </BootProvider>
  </StrictMode>,
);
