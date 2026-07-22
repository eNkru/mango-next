import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import { App } from './App';
import { BootProvider, routerBasename } from './lib/bootContext';
import { I18nProvider } from './lib/i18n';
import './styles/fonts.css';
import './styles/tokens.css';
import './styles/shell.css';

const rootEl = document.getElementById('root');
if (!rootEl) {
  throw new Error('#root missing');
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
