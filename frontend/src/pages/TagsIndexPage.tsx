import { useCallback, useEffect, useMemo, useState } from 'react';
import { apiFetch } from '../lib/api';
import { AppLink } from '../lib/AppLink';
import { useI18n } from '../lib/i18n';
import { AppShell } from '../shell/AppShell';
import { pushAlert } from '../shell/AlertHost';
import { Icon } from '../shell/Icon';
import { icons } from '../shell/icons';
import { EmptyState, ErrorState, LoadingState } from '../shell/StatePanels';

type TagInfo = {
  tag: string;
  count: number;
};

type TagsIndexResponse = {
  success?: boolean;
  tags?: TagInfo[];
};

export function TagsIndexPage() {
  const { t } = useI18n();
  const [tags, setTags] = useState<TagInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [query, setQuery] = useState('');

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await apiFetch<TagsIndexResponse>('api/tags/index');
      setTags(res.tags ?? []);
    } catch (err) {
      const message = err instanceof Error ? err.message : t('loadTagsFailed');
      setError(message);
      pushAlert(message, 'danger');
    } finally {
      setLoading(false);
    }
  }, [t]);

  useEffect(() => {
    void load();
  }, [load]);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return tags;
    return tags.filter((item) => item.tag.toLowerCase().includes(q));
  }, [query, tags]);

  return (
    <AppShell title={t('tags')} subtitle={t('tagsCount', { count: tags.length })}>
      {loading ? <LoadingState message={t('loading')} /> : null}
      {error ? (
        <ErrorState message={error} onRetry={() => void load()} retryLabel={t('retry')} />
      ) : null}

      {!loading && !error ? (
        <section className="mango-panel">
          <div className="mango-actions mango-actions--stack-sm">
            <input
              className="mango-input mango-max-w-search"
              type="search"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder={t('filterTags')}
              aria-label={t('filterTags')}
            />
            <button type="button" className="mango-btn" onClick={() => void load()}>
              <Icon icon={icons.refresh} size={16} />
              {t('refresh')}
            </button>
          </div>

          {filtered.length === 0 ? (
            <EmptyState message={query ? t('noMatchingTags') : t('noTagsYet')} />
          ) : (
            <div className="mango-tag-cloud">
              {filtered.map((item) => (
                <AppLink
                  key={item.tag}
                  className="mango-tag-pill"
                  to={`tags/${encodeURIComponent(item.tag)}`}
                >
                  {item.tag} <small>({item.count})</small>
                </AppLink>
              ))}
            </div>
          )}
        </section>
      ) : null}
    </AppShell>
  );
}
