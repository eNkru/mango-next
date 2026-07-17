import { useCallback, useEffect, useState } from 'react';
import { apiFetch } from '../lib/api';
import { baseUrl } from '../lib/baseUrl';
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
      const message = err instanceof Error ? err.message : '加载用户失败';
      setError(message);
      pushAlert(message, 'danger');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  const removeUser = async (username: string) => {
    setBusy(true);
    try {
      await apiFetch(`api/admin/user/delete/${encodeURIComponent(username)}`, {
        method: 'DELETE',
      });
      pushAlert(`已删除用户 ${username}`, 'success');
      setPendingDelete(null);
      await load();
    } catch (err) {
      pushAlert(err instanceof Error ? err.message : '删除失败', 'danger');
    } finally {
      setBusy(false);
    }
  };

  const empty = !loading && !error && users.length === 0;

  return (
    <AppShell title="用户管理" subtitle="创建、编辑和管理可登录用户">
      {loading ? <LoadingState message="正在加载用户…" /> : null}
      {error ? <ErrorState message={error} /> : null}
      {empty ? <EmptyState message="还没有用户" /> : null}

      {!loading && !error ? (
        <section className="mango-panel">
          <div className="mango-actions" style={{ marginTop: 0, marginBottom: '1rem' }}>
            <a className="mango-btn mango-btn--primary" href={baseUrl('admin/user/edit')}>
              新用户
            </a>
            <button type="button" className="mango-btn" disabled={busy} onClick={() => void load()}>
              刷新
            </button>
          </div>

          {users.length > 0 ? (
            <div style={{ overflowX: 'auto' }}>
              <table className="mango-table">
                <thead>
                  <tr>
                    <th>用户名</th>
                    <th>管理员权限</th>
                    <th>操作</th>
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
                            <span className="mango-badge" style={{ marginLeft: '0.5rem' }}>
                              当前
                            </span>
                          ) : null}
                        </td>
                        <td>
                          {user.admin ? (
                            <span className="mango-badge">是</span>
                          ) : (
                            <span className="mango-badge mango-badge--muted">否</span>
                          )}
                        </td>
                        <td>
                          <div className="mango-actions" style={{ marginTop: 0 }}>
                            <a
                              className="mango-btn"
                              href={baseUrl(
                                `admin/user/edit?username=${encodeURIComponent(user.username)}`,
                              )}
                            >
                              编辑
                            </a>
                            {!isSelf ? (
                              <button
                                type="button"
                                className="mango-btn mango-btn--danger"
                                disabled={busy}
                                onClick={() => setPendingDelete(user.username)}
                              >
                                删除
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
        title="确认删除用户？"
        message={
          pendingDelete
            ? `将删除用户 ${pendingDelete}。此操作不可撤销。`
            : ''
        }
        confirmLabel="删除"
        cancelLabel="取消"
        onCancel={() => setPendingDelete(null)}
        onConfirm={() => {
          if (pendingDelete) void removeUser(pendingDelete);
        }}
      />
    </AppShell>
  );
}
