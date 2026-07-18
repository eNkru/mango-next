import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { App } from './App';
import { I18nProvider } from './lib/i18n';
import './styles/tokens.css';
import './styles/shell.css';

const rootEl = document.getElementById('root');
if (!rootEl) {
  throw new Error('#root missing');
}

createRoot(rootEl).render(
  <StrictMode>
    <I18nProvider><App /></I18nProvider>
  </StrictMode>,
);
