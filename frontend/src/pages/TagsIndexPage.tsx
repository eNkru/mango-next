import { useCallback, useEffect, useMemo, useState } from 'react';
import { apiFetch } from '../lib/api';
import { baseUrl } from '../lib/baseUrl';
import { AppShell } from '../shell/AppShell';
import { pushAlert } from '../shell/AlertHost';
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
      const message = err instanceof Error ? err.message : '加载标签失败';
      setError(message);
      pushAlert(message, 'danger');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return tags;
    return tags.filter((t) => t.tag.toLowerCase().includes(q));
  }, [query, tags]);

  return (
    <AppShell title="标签" subtitle={`${tags.length} 个标签`}>
      {loading ? <LoadingState message="正在加载标签…" /> : null}
      {error ? <ErrorState message={error} /> : null}

      {!loading && !error ? (
        <section className="mango-panel">
          <div className="mango-actions" style={{ marginTop: 0, marginBottom: '1rem' }}>
            <input
              className="mango-input"
              style={{ maxWidth: '18rem' }}
              type="search"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="筛选标签…"
              aria-label="筛选标签"
            />
            <button type="button" className="mango-btn" onClick={() => void load()}>
              刷新
            </button>
          </div>

          {filtered.length === 0 ? (
            <EmptyState message={query ? '未找到匹配标签' : '还没有标签'} />
          ) : (
            <div className="mango-tag-cloud">
              {filtered.map((item) => (
                <a
                  key={item.tag}
                  className="mango-tag-pill"
                  href={baseUrl(`tags/${encodeURIComponent(item.tag)}`)}
                >
                  {item.tag} <small>({item.count})</small>
                </a>
              ))}
            </div>
          )}
        </section>
      ) : null}
    </AppShell>
  );
}
