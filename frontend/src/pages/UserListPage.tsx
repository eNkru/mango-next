import { useCallback, useEffect, useState } from 'react';
import { apiFetch } from '../lib/api';
import { baseUrl } from '../lib/baseUrl';
import { useI18n } from '../lib/i18n';
import { AppShell } from '../shell/AppShell';
import { ConfirmDialog } from '../shell/ConfirmDialog';
import { pushAlert } from '../shell/AlertHost';
import { EmptyState, ErrorState, LoadingState } from '../shell/StatePanels';

type UserRow = {
  username: string;
  admin: boolean;
};

type UsersResponse = {
  success?: boolean;
  users?: UserRow[];
  current_username?: string;
  error?: string;
};

export function UserListPage() {
  const { t } = useI18n();
  const [users, setUsers] = useState<UserRow[]>([]);
  const [currentUsername, setCurrentUsername] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [pendingDelete, setPendingDelete] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await apiFetch<UsersResponse>('api/admin/users');
      setUsers(res.users ?? []);
      setCurrentUsername(res.current_username ?? '');
    } catch (err) {
      const message = err instanceof Error ? err.message : t('loadUsersFailed');
      setError(message);
      pushAlert(message, 'danger');
    } finally {
      setLoading(false);
    }
  }, [t]);

  useEffect(() => {
    void load();
  }, [load]);

  const removeUser = async (username: string) => {
    setBusy(true);
    try {
      await apiFetch(`api/admin/user/delete/${encodeURIComponent(username)}`, {
        method: 'DELETE',
      });
      pushAlert(t('userDeleted', { username }), 'success');
      setPendingDelete(null);
      await load();
    } catch (err) {
      pushAlert(err instanceof Error ? err.message : t('deleteFailed'), 'danger');
    } finally {
      setBusy(false);
    }
  };

  const empty = !loading && !error && users.length === 0;

  return (
    <AppShell title={t('userManagement')} subtitle={t('userListSubtitle')}>
      {loading ? <LoadingState message={t('loadingUsers')} /> : null}
      {error ? (
        <ErrorState message={error} onRetry={() => void load()} retryLabel={t('retry')} />
      ) : null}
      {empty ? <EmptyState message={t('noUsersYet')} /> : null}

      {!loading && !error ? (
        <section className="mango-panel">
          <div className="mango-actions mango-actions--stack-sm">
            <a className="mango-btn mango-btn--primary" href={baseUrl('admin/user/edit')}>
              {t('newUser')}
            </a>
            <button type="button" className="mango-btn" disabled={busy} onClick={() => void load()}>
              {t('refresh')}
            </button>
          </div>

          {users.length > 0 ? (
            <div className="mango-scroll-x">
              <table className="mango-table">
                <thead>
                  <tr>
                    <th>{t('username')}</th>
                    <th>{t('adminPermission')}</th>
                    <th>{t('actions')}</th>
                  </tr>
                </thead>
                <tbody>
                  {users.map((user) => {
                    const isSelf = user.username === currentUsername;
                    return (
                      <tr key={user.username}>
                        <td>
                          <strong>{user.username}</strong>
                          {isSelf ? (
                            <span className="mango-badge mango-ml-2">{t('currentUser')}</span>
                          ) : null}
                        </td>
                        <td>
                          {user.admin ? (
                            <span className="mango-badge">{t('yes')}</span>
                          ) : (
                            <span className="mango-badge mango-badge--muted">{t('no')}</span>
                          )}
                        </td>
                        <td>
                          <div className="mango-actions mango-actions--flush">
                            <a
                              className="mango-btn"
                              href={baseUrl(
                                `admin/user/edit?username=${encodeURIComponent(user.username)}`,
                              )}
                            >
                              {t('edit')}
                            </a>
                            {!isSelf ? (
                              <button
                                type="button"
                                className="mango-btn mango-btn--danger"
                                disabled={busy}
                                onClick={() => setPendingDelete(user.username)}
                              >
                                {t('delete')}
                              </button>
                            ) : null}
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          ) : null}
        </section>
      ) : null}

      <ConfirmDialog
        open={pendingDelete !== null}
        title={t('confirmDeleteUser')}
        message={
          pendingDelete ? t('deleteUserMessage', { username: pendingDelete }) : ''
        }
        confirmLabel={t('delete')}
        cancelLabel={t('cancel')}
        onCancel={() => setPendingDelete(null)}
        onConfirm={() => {
          if (pendingDelete) void removeUser(pendingDelete);
        }}
      />
    </AppShell>
  );
}
