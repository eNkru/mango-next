import { FormEvent, useEffect, useMemo, useState } from 'react';
import { apiFetch } from '../lib/api';
import { baseUrl } from '../lib/baseUrl';
import { readBoot } from '../lib/boot';
import { AppShell } from '../shell/AppShell';
import { pushAlert } from '../shell/AlertHost';

export function UserEditPage() {
  const boot = useMemo(() => readBoot(), []);
  const originalUsername = useMemo(() => {
    if (boot.username) return boot.username;
    const params = new URLSearchParams(window.location.search);
    return params.get('username') ?? '';
  }, [boot.username]);
  const isNew = originalUsername === '';

  const [username, setUsername] = useState(originalUsername);
  const [password, setPassword] = useState('');
  const [admin, setAdmin] = useState(false);
  const [showPassword, setShowPassword] = useState(isNew);
  const [busy, setBusy] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const [loaded, setLoaded] = useState(isNew);

  useEffect(() => {
    if (isNew) return;
    let cancelled = false;
    void (async () => {
      try {
        const res = await apiFetch<{
          users?: Array<{ username: string; admin: boolean }>;
        }>('api/admin/users');
        if (cancelled) return;
        const match = (res.users ?? []).find((u) => u.username === originalUsername);
        if (match) {
          setAdmin(match.admin);
          setUsername(match.username);
        }
      } catch (err) {
        if (cancelled) return;
        const message = err instanceof Error ? err.message : '加载用户失败';
        setFormError(message);
        pushAlert(message, 'danger');
      } finally {
        if (!cancelled) setLoaded(true);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [isNew, originalUsername]);

  const onSubmit = async (event: FormEvent) => {
    event.preventDefault();
    setBusy(true);
    setFormError(null);
    try {
      if (isNew) {
        await apiFetch('api/admin/users', {
          method: 'POST',
          body: JSON.stringify({
            username,
            password,
            admin,
          }),
        });
        pushAlert('用户已创建', 'success');
      } else {
        await apiFetch(`api/admin/users/${encodeURIComponent(originalUsername)}`, {
          method: 'PUT',
          body: JSON.stringify({
            username,
            password: showPassword ? password : '',
            admin,
          }),
        });
        pushAlert('用户已更新', 'success');
      }
      window.location.href = baseUrl('admin/user');
    } catch (err) {
      const message = err instanceof Error ? err.message : '保存失败';
      setFormError(message);
      pushAlert(message, 'danger');
      setBusy(false);
    }
  };

  return (
    <AppShell
      title={isNew ? '新用户' : '编辑用户'}
      subtitle={isNew ? '创建可登录账户' : `编辑 ${originalUsername}`}
    >
      <section className="mango-panel">
        {!loaded ? <p>加载中…</p> : null}
        {loaded ? (
          <form className="mango-form" onSubmit={(e) => void onSubmit(e)}>
            <div className="mango-field">
              <label htmlFor="username">用户名</label>
              <input
                id="username"
                className="mango-input"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                required
                autoComplete="username"
              />
            </div>

            {isNew || showPassword ? (
              <div className="mango-field">
                <label htmlFor="password">{isNew ? '密码' : '新密码'}</label>
                <input
                  id="password"
                  className="mango-input"
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required={isNew}
                  autoComplete="new-password"
                />
              </div>
            ) : (
              <div className="mango-actions" style={{ marginTop: 0 }}>
                <button
                  type="button"
                  className="mango-btn"
                  onClick={() => setShowPassword(true)}
                >
                  更改密码
                </button>
              </div>
            )}

            <div className="mango-field mango-field--inline">
              <label htmlFor="admin">
                <input
                  id="admin"
                  type="checkbox"
                  checked={admin}
                  onChange={(e) => setAdmin(e.target.checked)}
                />{' '}
                管理员权限
              </label>
            </div>

            {formError ? (
              <p className="mango-state mango-state--error" role="alert">
                {formError}
              </p>
            ) : null}

            <div className="mango-actions">
              <button type="submit" className="mango-btn mango-btn--primary" disabled={busy}>
                {busy ? '保存中…' : '保存'}
              </button>
              <a className="mango-btn" href={baseUrl('admin/user')}>
                返回列表
              </a>
            </div>
          </form>
        ) : null}
      </section>
    </AppShell>
  );
}
