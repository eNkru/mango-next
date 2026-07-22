import type { ReactNode } from 'react';
import type { BrowseTitle, SortMode } from '../lib/browse';
import { AppLink } from '../lib/AppLink';
import { useI18n } from '../lib/i18n';
import { Icon } from '../shell/Icon';
import { icons } from '../shell/icons';

const ALL_MODES: SortMode[] = ['natural', 'title', 'modified', 'progress'];

export function ProgressBar({ value }: { value: number }) {
  const bounded = Math.max(0, Math.min(100, value || 0));
  return (
    <div className="mango-progress" aria-label={`${Math.round(bounded)}%`}>
      <span style={{ width: `${bounded}%` }} />
    </div>
  );
}

export function PosterCardSkeleton() {
  return (
    <div className="mango-card mango-card--skeleton" aria-hidden="true">
      <div className="mango-card__media mango-skeleton-shimmer" />
      <div className="mango-card__body">
        <div className="mango-skeleton-line mango-skeleton-line--title mango-skeleton-shimmer" />
        <div className="mango-skeleton-line mango-skeleton-line--meta mango-skeleton-shimmer" />
        <div className="mango-skeleton-line mango-skeleton-line--progress mango-skeleton-shimmer" />
      </div>
    </div>
  );
}

export function PosterCard({
  item,
  actions,
  showProgress = true,
}: {
  item: BrowseTitle;
  actions?: ReactNode;
  showProgress?: boolean;
}) {
  const { t } = useI18n();
  return (
    <article className={`mango-card${item.hidden ? ' mango-card--hidden' : ''}`}>
      <AppLink className="mango-card__link" to={`book/${encodeURIComponent(item.id)}`}>
        <div className="mango-card__media">
          {item.cover_url ? (
            <img src={item.cover_url} alt="" loading="lazy" />
          ) : (
            <div className="mango-card__placeholder" />
          )}
        </div>
        <div className="mango-card__body">
          <h3 className="mango-card__title">{item.name}</h3>
          <p className="mango-card__meta">
            {item.entry_count} {t('entries')}
          </p>
          {showProgress ? <ProgressBar value={item.progress} /> : null}
          {item.hidden ? <span className="mango-badge mango-badge--muted">{t('hidden')}</span> : null}
        </div>
      </AppLink>
      {actions ? <div className="mango-card__actions">{actions}</div> : null}
    </article>
  );
}

export function BrowseToolbar({
  query,
  onQuery,
  mode,
  onMode,
  ascending,
  onAscending,
  extra,
  modes = ALL_MODES,
}: {
  query: string;
  onQuery: (value: string) => void;
  mode: SortMode;
  onMode: (mode: SortMode) => void;
  ascending: boolean;
  onAscending: (value: boolean) => void;
  extra?: ReactNode;
  /** Subset of sort modes to offer; default is all four. */
  modes?: SortMode[];
}) {
  const { t } = useI18n();
  const allowed = modes.length ? modes : ALL_MODES;
  const activeMode = allowed.includes(mode) ? mode : allowed[0];
  const labels: Record<SortMode, string> = {
    natural: t('automatic'),
    title: t('title'),
    modified: t('modified'),
    progress: t('progress'),
  };

  return (
    <div className="mango-browse-toolbar">
      <label className="mango-browse-toolbar__search">
        <Icon icon={icons.search} size={16} />
        <span className="sr-only">{t('search')}</span>
        <input
          className="mango-input"
          type="search"
          value={query}
          onChange={(e) => onQuery(e.target.value)}
          placeholder={t('search')}
          aria-label={t('search')}
        />
      </label>
      <label>
        <span>{t('sort')}</span>
        <select
          className="mango-input"
          value={activeMode}
          onChange={(e) => onMode(e.target.value as SortMode)}
        >
          {allowed.map((item) => (
            <option key={item} value={item}>
              {labels[item]}
            </option>
          ))}
        </select>
      </label>
      <button
        type="button"
        className="mango-btn mango-btn--icon"
        onClick={() => onAscending(!ascending)}
        aria-label={ascending ? t('ascending') : t('descending')}
        title={ascending ? t('ascending') : t('descending')}
      >
        <Icon icon={ascending ? icons.sortAsc : icons.sortDesc} size={16} />
      </button>
      {extra}
    </div>
  );
}
