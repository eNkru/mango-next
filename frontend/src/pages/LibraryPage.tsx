import { useCallback, useEffect, useMemo, useState } from 'react';
import { apiFetch } from '../lib/api';
import {
  filterBrowseItems,
  sortBrowseItems,
  type BrowseTitle,
  type SortMode,
} from '../lib/browse';
import { useI18n } from '../lib/i18n';
import { BrowseToolbar, PosterCard } from '../browse/BrowseComponents';
import { AppShell } from '../shell/AppShell';
import { pushAlert } from '../shell/AlertHost';
import { Icon } from '../shell/Icon';
import { icons } from '../shell/icons';
import { EmptyState, ErrorState, LoadingState } from '../shell/StatePanels';

type LibraryResponse = { titles: BrowseTitle[]; is_admin: boolean; show_hidden: boolean };

function readState() {
  const params = new URLSearchParams(window.location.search);
  const mode = params.get('sort');
  return {
    query: params.get('q') ?? '',
    mode: (mode === 'title' || mode === 'modified' || mode === 'progress' ? mode : 'natural') as SortMode,
    ascending: params.get('order') !== 'desc',
  };
}

export function LibraryPage() {
  const { t } = useI18n();
  const initial = readState();
  const [query, setQuery] = useState(initial.query);
  const [mode, setMode] = useState<SortMode>(initial.mode);
  const [ascending, setAscending] = useState(initial.ascending);
  const [titles, setTitles] = useState<BrowseTitle[]>([]);
  const [isAdmin, setIsAdmin] = useState(false);
  const [showHidden, setShowHidden] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState<string | null>(null);

  const load = useCallback(async (hidden: boolean) => {
    setLoading(true);
    setError(null);
    try {
      const result = await apiFetch<LibraryResponse>(`api/library${hidden ? '?show_hidden=1' : ''}`);
      setTitles(result.titles ?? []);
      setIsAdmin(result.is_admin);
      setShowHidden(result.show_hidden);
    } catch (err) {
      const message = err instanceof Error ? err.message : t('loadFailed');
      setError(message);
      pushAlert(message, 'danger');
    } finally {
      setLoading(false);
    }
  }, [t]);

  useEffect(() => {
    void load(new URLSearchParams(window.location.search).get('show_hidden') === '1');
  }, [load]);

  useEffect(() => {
    const params = new URLSearchParams();
    if (query) params.set('q', query);
    if (mode !== 'natural') params.set('sort', mode);
    if (!ascending) params.set('order', 'desc');
    if (showHidden) params.set('show_hidden', '1');
    const value = params.toString();
    window.history.replaceState({}, '', `${window.location.pathname}${value ? `?${value}` : ''}`);
  }, [query, mode, ascending, showHidden]);

  const filtered = useMemo(
    () => sortBrowseItems(filterBrowseItems(titles, query), mode, ascending),
    [titles, query, mode, ascending],
  );

  const toggleHidden = async (item: BrowseTitle) => {
    setBusy(item.id);
    try {
      await apiFetch(`api/admin/hidden/${encodeURIComponent(item.id)}/${item.hidden ? 0 : 1}`, {
        method: 'PUT',
      });
      await load(showHidden);
    } catch (err) {
      pushAlert(err instanceof Error ? err.message : t('actionFailed'), 'danger');
    } finally {
      setBusy(null);
    }
  };

  return (
    <AppShell title={t('library')} subtitle={`${titles.length} ${t('entries')} · ${t('librarySubtitle')}`}>
      <BrowseToolbar
        query={query}
        onQuery={setQuery}
        mode={mode}
        onMode={setMode}
        ascending={ascending}
        onAscending={setAscending}
        extra={
          isAdmin ? (
            <button className="mango-btn" type="button" onClick={() => void load(!showHidden)}>
              <Icon icon={showHidden ? icons.hide : icons.show} size={16} />
              {showHidden ? t('hideHidden') : t('showHidden')}
            </button>
          ) : null
        }
      />
      {loading ? <LoadingState message={t('loading')} /> : null}
      {error ? (
        <ErrorState message={error} onRetry={() => void load(showHidden)} retryLabel={t('retry')} />
      ) : null}
      {!loading && !error && !filtered.length ? <EmptyState message={t('noResults')} /> : null}
      {!loading && !error && filtered.length ? (
        <div className="mango-card-grid">
          {filtered.map((item) => (
            <PosterCard
              key={item.id}
              item={item}
              actions={
                isAdmin ? (
                  <button
                    className="mango-btn mango-btn--danger"
                    type="button"
                    disabled={busy === item.id}
                    onClick={() => void toggleHidden(item)}
                  >
                    <Icon icon={item.hidden ? icons.show : icons.hide} size={16} />
                    {item.hidden ? t('show') : t('hide')}
                  </button>
                ) : undefined
              }
            />
          ))}
        </div>
      ) : null}
    </AppShell>
  );
}
