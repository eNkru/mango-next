import { FormEvent, useMemo, useState } from 'react';
import { apiFetch } from '../lib/api';
import { readBoot } from '../lib/boot';
import { resolvePostLoginHref, safeRedirectPath } from '../lib/safeRedirect';

type LoginResponse = {
  success?: boolean;
  session_id?: string;
  is_admin?: boolean;
  error?: string;
};

export function LoginPage() {
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
      setError('登录失败，请检查用户名和密码');
      setBusy(false);
    }
  };

  return (
    <div className="mango-login">
      <div className="mango-login__card">
        <header className="mango-login__header">
          <h1>欢迎回来</h1>
          <p>登录到 Mango</p>
        </header>
        <form className="mango-login__form" onSubmit={(e) => void onSubmit(e)}>
          {error ? (
            <div className="mango-login__error" role="alert">
              {error}
            </div>
          ) : null}
          <div className="mango-field">
            <label htmlFor="username">用户名</label>
            <input
              id="username"
              className="mango-input"
              name="username"
              type="text"
              autoComplete="username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              disabled={busy}
              placeholder="请输入用户名"
            />
          </div>
          <div className="mango-field">
            <label htmlFor="password">密码</label>
            <div className="mango-login__password-row">
              <input
                id="password"
                className="mango-input"
                name="password"
                type={showPassword ? 'text' : 'password'}
                autoComplete="current-password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                disabled={busy}
                placeholder="请输入密码"
              />
              <button
                type="button"
                className="mango-login__toggle"
                onClick={() => setShowPassword((v) => !v)}
                aria-label={showPassword ? '隐藏密码' : '显示密码'}
                disabled={busy}
              >
                {showPassword ? '隐藏' : '显示'}
              </button>
            </div>
          </div>
          <button
            className="mango-btn mango-btn--primary mango-login__submit"
            type="submit"
            disabled={busy || !username || !password}
          >
            {busy ? '登录中…' : '登录'}
          </button>
        </form>
        <footer className="mango-login__footer">
          <p>Mango · 漫画服务器</p>
        </footer>
      </div>
    </div>
  );
}
