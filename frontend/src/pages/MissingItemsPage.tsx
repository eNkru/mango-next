import { useCallback, useEffect, useState } from 'react';
import { apiFetch } from '../lib/api';
import { useI18n } from '../lib/i18n';
import { AppShell } from '../shell/AppShell';
import { ConfirmDialog } from '../shell/ConfirmDialog';
import { pushAlert } from '../shell/AlertHost';
import { Icon } from '../shell/Icon';
import { icons } from '../shell/icons';
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
  const { t } = useI18n();
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
      const message = err instanceof Error ? err.message : t('loadFailed');
      setError(message);
      pushAlert(message, 'danger');
    } finally {
      setLoading(false);
    }
  }, [t]);

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
      pushAlert(t('deleted'), 'success');
      await load();
    } catch (err) {
      pushAlert(err instanceof Error ? err.message : t('deleteFailed'), 'danger');
    } finally {
      setBusy(false);
    }
  };

  const removeAll = async () => {
    setBusy(true);
    try {
      await apiFetch('api/admin/titles/missing', { method: 'DELETE' });
      await apiFetch('api/admin/entries/missing', { method: 'DELETE' });
      pushAlert(t('deletedAllMissing'), 'success');
      await load();
    } catch (err) {
      pushAlert(err instanceof Error ? err.message : t('bulkDeleteFailed'), 'danger');
    } finally {
      setBusy(false);
      setConfirmBulk(false);
    }
  };

  const empty = !loading && !error && titles.length === 0 && entries.length === 0;

  return (
    <AppShell title={t('missingTitle')} subtitle={t('missingSubtitle')}>
      {loading ? <LoadingState message={t('loadingMissing')} /> : null}
      {error ? (
        <ErrorState message={error} onRetry={() => void load()} retryLabel={t('retry')} />
      ) : null}
      {empty ? <EmptyState message={t('noMissingItems')} /> : null}

      {!loading && !error && !empty ? (
        <section className="mango-panel">
          <p className="mango-muted-copy">{t('missingHelp')}</p>
          <div className="mango-actions">
            <button
              type="button"
              className="mango-btn mango-btn--danger"
              disabled={busy}
              onClick={() => setConfirmBulk(true)}
            >
              <Icon icon={icons.delete} size={16} />
              {t('deleteAll')}
            </button>
            <button type="button" className="mango-btn" disabled={busy} onClick={() => void load()}>
              <Icon icon={icons.refresh} size={16} />
              {t('refresh')}
            </button>
          </div>

          <div className="mango-scroll-x mango-mt-1">
            <table className="mango-table">
              <thead>
                <tr>
                  <th>{t('type')}</th>
                  <th>{t('relativePath')}</th>
                  <th>{t('id')}</th>
                  <th>{t('actions')}</th>
                </tr>
              </thead>
              <tbody>
                {titles.map((item) => (
                  <tr key={`title-${item.id}`}>
                    <td>
                      <span className="mango-badge">{t('titleKind')}</span>
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
                        <Icon icon={icons.delete} size={16} />
                        {t('delete')}
                      </button>
                    </td>
                  </tr>
                ))}
                {entries.map((item) => (
                  <tr key={`entry-${item.id}`}>
                    <td>
                      <span className="mango-badge mango-badge--muted">{t('pathKind')}</span>
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
                        <Icon icon={icons.delete} size={16} />
                        {t('delete')}
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
        title={t('confirmDeleteAll')}
        message={t('confirmDeleteAllMessage')}
        confirmLabel={t('confirmDeleteAllYes')}
        cancelLabel={t('cancel')}
        onCancel={() => setConfirmBulk(false)}
        onConfirm={() => void removeAll()}
      />
    </AppShell>
  );
}
