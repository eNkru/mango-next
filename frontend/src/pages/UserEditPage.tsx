import { FormEvent, useEffect, useMemo, useState } from 'react';
import { apiFetch } from '../lib/api';
import { baseUrl } from '../lib/baseUrl';
import { readBoot } from '../lib/boot';
import { useI18n } from '../lib/i18n';
import { AppShell } from '../shell/AppShell';
import { pushAlert } from '../shell/AlertHost';
import { Icon } from '../shell/Icon';
import { icons } from '../shell/icons';
import { ErrorState, LoadingState } from '../shell/StatePanels';

export function UserEditPage() {
  const { t } = useI18n();
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
  const [loadError, setLoadError] = useState<string | null>(null);
  const [loaded, setLoaded] = useState(isNew);
  const [loadNonce, setLoadNonce] = useState(0);

  useEffect(() => {
    if (isNew) return;
    let cancelled = false;
    setLoaded(false);
    setLoadError(null);
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
        const message = err instanceof Error ? err.message : t('loadUserFailed');
        setLoadError(message);
        pushAlert(message, 'danger');
      } finally {
        if (!cancelled) setLoaded(true);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [isNew, originalUsername, t, loadNonce]);

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
        pushAlert(t('userCreated'), 'success');
      } else {
        await apiFetch(`api/admin/users/${encodeURIComponent(originalUsername)}`, {
          method: 'PUT',
          body: JSON.stringify({
            username,
            password: showPassword ? password : '',
            admin,
          }),
        });
        pushAlert(t('userUpdated'), 'success');
      }
      window.location.href = baseUrl('admin/user');
    } catch (err) {
      const message = err instanceof Error ? err.message : t('saveFailed');
      setFormError(message);
      pushAlert(message, 'danger');
      setBusy(false);
    }
  };

  return (
    <AppShell
      title={isNew ? t('newUser') : t('editUser')}
      subtitle={isNew ? t('createAccount') : t('editUserSubtitle', { username: originalUsername })}
    >
      <section className="mango-panel">
        {!loaded ? <LoadingState message={t('loading')} /> : null}
        {loaded && loadError ? (
          <ErrorState
            message={loadError}
            onRetry={() => setLoadNonce((n) => n + 1)}
            retryLabel={t('retry')}
          />
        ) : null}
        {loaded && !loadError ? (
          <form className="mango-form" onSubmit={(e) => void onSubmit(e)}>
            <label className="mango-field">
              <span>{t('username')}</span>
              <input
                className="mango-input"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                required
                autoComplete="username"
              />
            </label>

            {isNew || showPassword ? (
              <label className="mango-field">
                <span>{isNew ? t('password') : t('newPassword')}</span>
                <input
                  className="mango-input"
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required={isNew}
                  autoComplete="new-password"
                />
              </label>
            ) : (
              <div className="mango-actions mango-actions--flush">
                <button type="button" className="mango-btn" onClick={() => setShowPassword(true)}>
                  <Icon icon={icons.edit} size={16} />
                  {t('changePassword')}
                </button>
              </div>
            )}

            <label className="mango-field mango-field--inline">
              <input
                type="checkbox"
                checked={admin}
                onChange={(e) => setAdmin(e.target.checked)}
              />
              <span>{t('adminPermission')}</span>
            </label>

            {formError ? <ErrorState message={formError} /> : null}

            <div className="mango-actions">
              <button type="submit" className="mango-btn mango-btn--primary" disabled={busy}>
                <Icon icon={icons.save} size={16} />
                {busy ? t('saving') : t('save')}
              </button>
              <a className="mango-btn" href={baseUrl('admin/user')}>
                <Icon icon={icons.back} size={16} />
                {t('backToList')}
              </a>
            </div>
          </form>
        ) : null}
      </section>
    </AppShell>
  );
}
