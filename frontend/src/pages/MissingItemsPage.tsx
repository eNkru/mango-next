import { useCallback, useEffect, useState } from 'react';
import { apiFetch } from '../lib/api';
import { AppShell } from '../shell/AppShell';
import { ConfirmDialog } from '../shell/ConfirmDialog';
import { pushAlert } from '../shell/AlertHost';
import { EmptyState, ErrorState, LoadingState } from '../shell/StatePanels';

type MissingItem = {
  id: string;
  path: string;
};

type TitlesResponse = {
  success?: boolean;
  titles?: MissingItem[];
  error?: string;
};

type EntriesResponse = {
  success?: boolean;
  entries?: MissingItem[];
  error?: string;
};

export function MissingItemsPage() {
  const [titles, setTitles] = useState<MissingItem[]>([]);
  const [entries, setEntries] = useState<MissingItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [confirmBulk, setConfirmBulk] = useState(false);
  const [busy, setBusy] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const titlesRes = await apiFetch<TitlesResponse>('api/admin/titles/missing');
      const entriesRes = await apiFetch<EntriesResponse>('api/admin/entries/missing');
      setTitles(titlesRes.titles ?? []);
      setEntries(entriesRes.entries ?? []);
    } catch (err) {
      const message = err instanceof Error ? err.message : '加载失败';
      setError(message);
      pushAlert(message, 'danger');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  const removeOne = async (kind: 'title' | 'entry', id: string) => {
    setBusy(true);
    try {
      const path =
        kind === 'title'
          ? `api/admin/titles/missing/${encodeURIComponent(id)}`
          : `api/admin/entries/missing/${encodeURIComponent(id)}`;
      await apiFetch(path, { method: 'DELETE' });
      pushAlert('已删除', 'success');
      await load();
    } catch (err) {
      pushAlert(err instanceof Error ? err.message : '删除失败', 'danger');
    } finally {
      setBusy(false);
    }
  };

  const removeAll = async () => {
    setBusy(true);
    try {
      await apiFetch('api/admin/titles/missing', { method: 'DELETE' });
      await apiFetch('api/admin/entries/missing', { method: 'DELETE' });
      pushAlert('已删除全部缺失项', 'success');
      await load();
    } catch (err) {
      pushAlert(err instanceof Error ? err.message : '批量删除失败', 'danger');
    } finally {
      setBusy(false);
      setConfirmBulk(false);
    }
  };

  const empty = !loading && !error && titles.length === 0 && entries.length === 0;

  return (
    <AppShell
      title="缺失条目"
      subtitle="资料库中记录存在，但磁盘上已找不到对应文件的项目"
    >
      {loading ? <LoadingState message="正在加载缺失条目…" /> : null}
      {error ? <ErrorState message={error} /> : null}
      {empty ? <EmptyState message="没有找到丢失的条目，所有条目均正常" /> : null}

      {!loading && !error && !empty ? (
        <section className="mango-panel">
          <p style={{ color: 'var(--mango-text-muted)', lineHeight: 1.6 }}>
            以下项目存在于资料库元数据中，但现在找不到对应文件。若误删，请恢复文件后重新扫描；
            否则可删除元数据以释放数据库空间。
          </p>
          <div className="mango-actions">
            <button
              type="button"
              className="mango-btn mango-btn--danger"
              disabled={busy}
              onClick={() => setConfirmBulk(true)}
            >
              删除全部
            </button>
            <button type="button" className="mango-btn" disabled={busy} onClick={() => void load()}>
              刷新
            </button>
          </div>

          <div style={{ overflowX: 'auto', marginTop: '1rem' }}>
            <table className="mango-table">
              <thead>
                <tr>
                  <th>类型</th>
                  <th>相对路径</th>
                  <th>ID</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {titles.map((item) => (
                  <tr key={`title-${item.id}`}>
                    <td>
                      <span className="mango-badge">标题</span>
                    </td>
                    <td>{item.path}</td>
                    <td>
                      <code>{item.id}</code>
                    </td>
                    <td>
                      <button
                        type="button"
                        className="mango-btn mango-btn--danger"
                        disabled={busy}
                        onClick={() => void removeOne('title', item.id)}
                      >
                        删除
                      </button>
                    </td>
                  </tr>
                ))}
                {entries.map((item) => (
                  <tr key={`entry-${item.id}`}>
                    <td>
                      <span className="mango-badge mango-badge--muted">路径</span>
                    </td>
                    <td>{item.path}</td>
                    <td>
                      <code>{item.id}</code>
                    </td>
                    <td>
                      <button
                        type="button"
                        className="mango-btn mango-btn--danger"
                        disabled={busy}
                        onClick={() => void removeOne('entry', item.id)}
                      >
                        删除
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>
      ) : null}

      <ConfirmDialog
        open={confirmBulk}
        title="确认删除全部？"
        message="与这些项目相关的所有元数据，包括标签和缩略图，都将从数据库中删除。"
        confirmLabel="是的，删除它们"
        cancelLabel="取消"
        onCancel={() => setConfirmBulk(false)}
        onConfirm={() => void removeAll()}
      />
    </AppShell>
  );
}
