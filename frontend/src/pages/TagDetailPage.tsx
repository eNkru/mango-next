import { useCallback, useEffect, useMemo, useState } from 'react';
import { apiFetch } from '../lib/api';
import { baseUrl } from '../lib/baseUrl';
import { readBoot } from '../lib/boot';
import { AppShell } from '../shell/AppShell';
import { pushAlert } from '../shell/AlertHost';
import { EmptyState, ErrorState, LoadingState } from '../shell/StatePanels';

type TitleCard = {
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
  titles?: TitleCard[];
};

export function TagDetailPage() {
  const boot = useMemo(() => readBoot(), []);
  const tag = useMemo(() => {
    if (boot.tag) return boot.tag;
    const parts = window.location.pathname.split('/').filter(Boolean);
    const idx = parts.lastIndexOf('tags');
    if (idx >= 0 && parts[idx + 1]) return decodeURIComponent(parts[idx + 1]);
    return '';
  }, [boot.tag]);

  const [titles, setTitles] = useState<TitleCard[]>([]);
  const [isAdmin, setIsAdmin] = useState(Boolean(boot.isAdmin));
  const [showHidden, setShowHidden] = useState(Boolean(boot.showHidden));
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [query, setQuery] = useState('');
  const [busyId, setBusyId] = useState<string | null>(null);

  const load = useCallback(async (hidden: boolean) => {
    if (!tag) {
      setError('缺少标签');
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
      setTitles(res.titles ?? []);
      setIsAdmin(Boolean(res.is_admin));
      setShowHidden(Boolean(res.show_hidden));
    } catch (err) {
      const message = err instanceof Error ? err.message : '加载标签失败';
      setError(message);
      pushAlert(message, 'danger');
    } finally {
      setLoading(false);
    }
  }, [tag]);

  useEffect(() => {
    void load(Boolean(boot.showHidden));
  }, [boot.showHidden, load]);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return titles;
    return titles.filter((t) => t.name.toLowerCase().includes(q));
  }, [query, titles]);

  const toggleShowHidden = () => {
    const next = !showHidden;
    const params = new URLSearchParams(window.location.search);
    if (next) params.set('show_hidden', '1');
    else params.delete('show_hidden');
    const qs = params.toString();
    window.history.replaceState({}, '', `${window.location.pathname}${qs ? `?${qs}` : ''}`);
    void load(next);
  };

  const toggleHidden = async (tid: string, value: 0 | 1) => {
    setBusyId(tid);
    try {
      await apiFetch(`api/admin/hidden/${encodeURIComponent(tid)}/${value}`, {
        method: 'PUT',
      });
      pushAlert(value === 1 ? '已隐藏' : '已显示', 'success');
      await load(showHidden);
    } catch (err) {
      pushAlert(err instanceof Error ? err.message : '操作失败', 'danger');
    } finally {
      setBusyId(null);
    }
  };

  return (
    <AppShell title={`标签: ${tag || '…'}`} subtitle={`${titles.length} 个标记`}>
      <div className="mango-actions" style={{ marginTop: 0, marginBottom: '1rem' }}>
        <a className="mango-btn" href={baseUrl('tags')}>
          全部标签
        </a>
        <input
          className="mango-input"
          style={{ maxWidth: '16rem' }}
          type="search"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="查找漫画…"
          aria-label="查找漫画"
        />
        {isAdmin ? (
          <button type="button" className="mango-btn" onClick={toggleShowHidden}>
            {showHidden ? '隐藏已隐藏' : '显示隐藏'}
          </button>
        ) : null}
      </div>

      {loading ? <LoadingState message="正在加载…" /> : null}
      {error ? <ErrorState message={error} /> : null}
      {!loading && !error && filtered.length === 0 ? (
        <EmptyState message={query ? '未找到结果' : '该标签下没有漫画'} />
      ) : null}

      {!loading && !error && filtered.length > 0 ? (
        <div className="mango-card-grid">
          {filtered.map((item) => (
            <article
              key={item.id}
              className={`mango-card${item.hidden && showHidden ? ' mango-card--hidden' : ''}`}
            >
              <a className="mango-card__link" href={baseUrl(`book/${encodeURIComponent(item.id)}`)}>
                <div className="mango-card__media">
                  {item.cover_url ? (
                    <img src={item.cover_url} alt="" loading="lazy" />
                  ) : (
                    <div className="mango-card__placeholder" />
                  )}
                </div>
                <div className="mango-card__body">
                  <h3 className="mango-card__title">{item.name}</h3>
                  <p className="mango-card__meta">{item.entry_count} 项</p>
                  {item.hidden && showHidden ? (
                    <span className="mango-badge mango-badge--muted">已隐藏</span>
                  ) : null}
                </div>
              </a>
              {isAdmin ? (
                <div className="mango-card__actions">
                  <button
                    type="button"
                    className="mango-btn mango-btn--danger"
                    disabled={busyId === item.id}
                    onClick={() => void toggleHidden(item.id, item.hidden ? 0 : 1)}
                  >
                    {item.hidden ? '显示' : '隐藏'}
                  </button>
                </div>
              ) : null}
            </article>
          ))}
        </div>
      ) : null}
    </AppShell>
  );
}
