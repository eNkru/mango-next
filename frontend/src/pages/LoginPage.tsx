import { FormEvent, useMemo, useState } from 'react';
import { apiFetch } from '../lib/api';
import { readBoot } from '../lib/boot';
import { useI18n } from '../lib/i18n';
import { resolvePostLoginHref, safeRedirectPath } from '../lib/safeRedirect';
import { LanguageSelect } from '../shell/LanguageSelect';

type LoginResponse = {
  success?: boolean;
  session_id?: string;
  is_admin?: boolean;
  error?: string;
};

export function LoginPage() {
  const { t } = useI18n();
  const boot = useMemo(() => readBoot(), []);
  const callback = useMemo(() => {
    if (boot.callback) return safeRedirectPath(boot.callback);
    const params = new URLSearchParams(window.location.search);
    return safeRedirectPath(params.get('callback'));
  }, [boot.callback]);

  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const onSubmit = async (event: FormEvent) => {
    event.preventDefault();
    setBusy(true);
    setError(null);
    try {
      await apiFetch<LoginResponse>('api/login', {
        method: 'POST',
        body: JSON.stringify({ username, password }),
      });
      window.location.assign(resolvePostLoginHref(boot.baseUrl, callback));
    } catch {
      setError(t('loginFailed'));
      setBusy(false);
    }
  };

  return (
    <div className="mango-login">
      <div className="mango-login__card">
        <header className="mango-login__header">
          <h1>{t('loginWelcome')}</h1>
          <p>{t('loginSubtitle')}</p>
        </header>
        <form className="mango-login__form" onSubmit={(e) => void onSubmit(e)}>
          {error ? (
            <div className="mango-login__error" role="alert">
              {error}
            </div>
          ) : null}
          <label className="mango-field">
            <span>{t('username')}</span>
            <input
              className="mango-input"
              name="username"
              type="text"
              autoComplete="username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              disabled={busy}
              placeholder={t('usernamePlaceholder')}
            />
          </label>
          <label className="mango-field">
            <span>{t('password')}</span>
            <div className="mango-login__password-row">
              <input
                className="mango-input"
                name="password"
                type={showPassword ? 'text' : 'password'}
                autoComplete="current-password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                disabled={busy}
                placeholder={t('passwordPlaceholder')}
              />
              <button
                type="button"
                className="mango-login__toggle"
                onClick={() => setShowPassword((v) => !v)}
                aria-label={showPassword ? t('hidePassword') : t('showPassword')}
                disabled={busy}
              >
                {showPassword ? t('hidePassword') : t('showPassword')}
              </button>
            </div>
          </label>
          <button
            className="mango-btn mango-btn--primary mango-login__submit"
            type="submit"
            disabled={busy || !username || !password}
          >
            {busy ? t('loggingIn') : t('login')}
          </button>
        </form>
        <footer className="mango-login__footer">
          <LanguageSelect />
          <p>{t('loginFooter')}</p>
        </footer>
      </div>
    </div>
  );
}
