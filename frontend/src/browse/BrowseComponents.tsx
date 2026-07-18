import type { ReactNode } from 'react';
import type { BrowseTitle, SortMode } from '../lib/browse';
import { baseUrl } from '../lib/baseUrl';
import { useI18n } from '../lib/i18n';

export function ProgressBar({ value }: { value: number }) {
  const bounded = Math.max(0, Math.min(100, value || 0));
  return <div className="mango-progress" aria-label={`${Math.round(bounded)}%`}><span style={{ width: `${bounded}%` }} /></div>;
}

export function PosterCard({ item, actions }: { item: BrowseTitle; actions?: ReactNode }) {
  const { t } = useI18n();
  return <article className={`mango-card${item.hidden ? ' mango-card--hidden' : ''}`}>
    <a className="mango-card__link" href={baseUrl(`book/${encodeURIComponent(item.id)}`)}>
      <div className="mango-card__media">{item.cover_url ? <img src={item.cover_url} alt="" loading="lazy" /> : <div className="mango-card__placeholder" />}</div>
      <div className="mango-card__body">
        <h3 className="mango-card__title">{item.name}</h3>
        <p className="mango-card__meta">{item.entry_count} {t('entries')}</p>
        <ProgressBar value={item.progress} />
        {item.hidden ? <span className="mango-badge mango-badge--muted">{t('hidden')}</span> : null}
      </div>
    </a>
    {actions ? <div className="mango-card__actions">{actions}</div> : null}
  </article>;
}

export function BrowseToolbar({ query, onQuery, mode, onMode, ascending, onAscending, extra }: {
  query: string; onQuery: (value: string) => void; mode: SortMode; onMode: (mode: SortMode) => void;
  ascending: boolean; onAscending: (value: boolean) => void; extra?: ReactNode;
}) {
  const { t } = useI18n();
  return <div className="mango-browse-toolbar">
    <input className="mango-input" type="search" value={query} onChange={(e) => onQuery(e.target.value)} placeholder={t('search')} aria-label={t('search')} />
    <label><span>{t('sort')}</span><select className="mango-input" value={mode} onChange={(e) => onMode(e.target.value as SortMode)}>
      <option value="natural">{t('automatic')}</option><option value="title">{t('title')}</option>
      <option value="modified">{t('modified')}</option><option value="progress">{t('progress')}</option>
    </select></label>
    <label><span className="sr-only">{t('sort')}</span><select className="mango-input" value={ascending ? 'asc' : 'desc'} onChange={(e) => onAscending(e.target.value === 'asc')}>
      <option value="asc">{t('ascending')}</option><option value="desc">{t('descending')}</option>
    </select></label>{extra}
  </div>;
}
