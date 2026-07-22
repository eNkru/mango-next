import { useCallback, useEffect, useMemo, useState } from 'react';
import { BrowseToolbar, PosterCard } from '../browse/BrowseComponents';
import { apiFetch } from '../lib/api';
import { AppLink } from '../lib/AppLink';
import { useBoot } from '../lib/bootContext';
import {
  filterBrowseItems,
  sortBrowseItems,
  type BrowseTitle,
  type SortMode,
} from '../lib/browse';
import { useI18n } from '../lib/i18n';
import { AppShell } from '../shell/AppShell';
import { pushAlert } from '../shell/AlertHost';
import { Icon } from '../shell/Icon';
import { icons } from '../shell/icons';
import { EmptyState, ErrorState, LoadingState } from '../shell/StatePanels';

type TagApiTitle = {
  id: string;
  name: string;
  cover_url: string;
  entry_count: number;
  hidden: boolean;
};

type TagDetailResponse = {
  success?: boolean;
  tag?: string;
  show_hidden?: boolean;
  is_admin?: boolean;
  titles?: TagApiTitle[];
};

const TAG_SORT_MODES: SortMode[] = ['natural', 'title'];

function toBrowseTitle(card: TagApiTitle): BrowseTitle {
  return {
    id: card.id,
    name: card.name,
    display_name: card.name,
    file_name: card.name,
    sort_name: card.name,
    cover_url: card.cover_url,
    entry_count: card.entry_count,
    progress: 0,
    modified_at: 0,
    hidden: card.hidden,
  };
}

export function TagDetailPage({ tag, showHidden: initialShowHidden = false }: { tag: string; showHidden?: boolean }) {
  const { t } = useI18n();
  const boot = useBoot();

  const [titles, setTitles] = useState<BrowseTitle[]>([]);
  const [isAdmin, setIsAdmin] = useState(Boolean(boot.isAdmin));
  const [showHidden, setShowHidden] = useState(Boolean(initialShowHidden));
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [query, setQuery] = useState('');
  const [mode, setMode] = useState<SortMode>('natural');
  const [ascending, setAscending] = useState(true);
  const [busyId, setBusyId] = useState<string | null>(null);

  const load = useCallback(
    async (hidden: boolean) => {
      if (!tag) {
        setError(t('missingTag'));
        setLoading(false);
        return;
      }
      setLoading(true);
      setError(null);
      try {
        const qs = hidden ? '?show_hidden=1' : '';
        const res = await apiFetch<TagDetailResponse>(
          `api/tags/${encodeURIComponent(tag)}/titles${qs}`,
        );
        setTitles((res.titles ?? []).map(toBrowseTitle));
        setIsAdmin(Boolean(res.is_admin));
        setShowHidden(Boolean(res.show_hidden));
      } catch (err) {
        const message = err instanceof Error ? err.message : t('loadTagFailed');
        setError(message);
        pushAlert(message, 'danger');
      } finally {
        setLoading(false);
      }
    },
    [tag, t],
  );

  useEffect(() => {
    setShowHidden(Boolean(initialShowHidden));
    void load(Boolean(initialShowHidden));
  }, [initialShowHidden, load, tag]);

  const filtered = useMemo(
    () => sortBrowseItems(filterBrowseItems(titles, query), mode, ascending),
    [titles, query, mode, ascending],
  );

  const toggleShowHidden = () => {
    const next = !showHidden;
    const params = new URLSearchParams(window.location.search);
    if (next) params.set('show_hidden', '1');
    else params.delete('show_hidden');
    const qs = params.toString();
    window.history.replaceState({}, '', `${window.location.pathname}${qs ? `?${qs}` : ''}`);
    void load(next);
  };

  const toggleHidden = async (item: BrowseTitle) => {
    setBusyId(item.id);
    try {
      await apiFetch(`api/admin/hidden/${encodeURIComponent(item.id)}/${item.hidden ? 0 : 1}`, {
        method: 'PUT',
      });
      pushAlert(item.hidden ? t('shownDone') : t('hiddenDone'), 'success');
      await load(showHidden);
    } catch (err) {
      pushAlert(err instanceof Error ? err.message : t('actionFailed'), 'danger');
    } finally {
      setBusyId(null);
    }
  };

  return (
    <AppShell
      title={t('tagTitle', { tag: tag || '…' })}
      subtitle={t('tagCount', { count: titles.length })}
    >
      <BrowseToolbar
        query={query}
        onQuery={setQuery}
        mode={mode}
        onMode={setMode}
        ascending={ascending}
        onAscending={setAscending}
        modes={TAG_SORT_MODES}
        extra={
          <>
            <AppLink className="mango-btn" to="tags">
              <Icon icon={icons.back} size={16} />
              {t('allTags')}
            </AppLink>
            {isAdmin ? (
              <button type="button" className="mango-btn" onClick={toggleShowHidden}>
                <Icon icon={showHidden ? icons.hide : icons.show} size={16} />
                {showHidden ? t('hideHidden') : t('showHidden')}
              </button>
            ) : null}
          </>
        }
      />

      {loading ? <LoadingState message={t('loading')} /> : null}
      {error ? (
        <ErrorState message={error} onRetry={() => void load(showHidden)} retryLabel={t('retry')} />
      ) : null}
      {!loading && !error && filtered.length === 0 ? (
        <EmptyState message={query ? t('noResults') : t('noMangaInTag')} />
      ) : null}

      {!loading && !error && filtered.length > 0 ? (
        <div className="mango-card-grid">
          {filtered.map((item) => (
            <PosterCard
              key={item.id}
              item={item}
              showProgress={false}
              actions={
                isAdmin ? (
                  <button
                    type="button"
                    className="mango-btn mango-btn--danger"
                    disabled={busyId === item.id}
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
