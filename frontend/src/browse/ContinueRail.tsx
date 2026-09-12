import { useCallback, useEffect, useRef, useState } from 'react';
import type { BrowseEntry } from '../lib/browse';
import { AppLink } from '../lib/AppLink';
import { useI18n } from '../lib/i18n';
import { ProgressBar } from './BrowseComponents';
import { continueReaderPath } from './continueReaderPath';
import { Icon } from '../shell/Icon';
import { icons } from '../shell/icons';

export function ContinueRail({ items }: { items: BrowseEntry[] }) {
  const { t } = useI18n();
  const trackRef = useRef<HTMLDivElement>(null);
  const [canPrev, setCanPrev] = useState(false);
  const [canNext, setCanNext] = useState(false);

  const updateEdges = useCallback(() => {
    const track = trackRef.current;
    if (!track) {
      setCanPrev(false);
      setCanNext(false);
      return;
    }
    const max = track.scrollWidth - track.clientWidth;
    const left = track.scrollLeft;
    setCanPrev(left > 2);
    setCanNext(max > 2 && left < max - 2);
  }, []);

  useEffect(() => {
    const track = trackRef.current;
    if (!track) return;
    updateEdges();
    const onScroll = () => updateEdges();
    track.addEventListener('scroll', onScroll, { passive: true });
    const ro = typeof ResizeObserver !== 'undefined' ? new ResizeObserver(updateEdges) : null;
    ro?.observe(track);
    window.addEventListener('resize', updateEdges);
    return () => {
      track.removeEventListener('scroll', onScroll);
      ro?.disconnect();
      window.removeEventListener('resize', updateEdges);
    };
  }, [updateEdges, items.length]);

  const scrollByPage = (dir: -1 | 1) => {
    const track = trackRef.current;
    if (!track) return;
    const amount = Math.max(track.clientWidth * 0.85, 180);
    track.scrollBy({ left: dir * amount, behavior: 'smooth' });
  };

  return (
    <section className="mango-browse-section">
      <h2>{t('continueReading')}</h2>
      <div className="mango-poster-rail-shell">
        {canPrev ? (
          <button
            type="button"
            className="mango-poster-rail__arrow mango-poster-rail__arrow--prev mango-btn mango-btn--icon"
            aria-label={t('previousEntry')}
            onClick={() => scrollByPage(-1)}
          >
            <Icon icon={icons.back} size={18} />
          </button>
        ) : null}
        <div ref={trackRef} className="mango-continue-rail">
          {items.map((item) => (
            <ContinueCard key={item.id} item={item} />
          ))}
        </div>
        {canNext ? (
          <button
            type="button"
            className="mango-poster-rail__arrow mango-poster-rail__arrow--next mango-btn mango-btn--icon"
            aria-label={t('nextEntry')}
            onClick={() => scrollByPage(1)}
          >
            <Icon icon={icons.forward} size={18} />
          </button>
        ) : null}
      </div>
    </section>
  );
}

function ContinueCard({ item }: { item: BrowseEntry }) {
  const { t } = useI18n();
  const to = continueReaderPath(item);

  return (
    <AppLink className="mango-continue-rail__card" to={to}>
      <div className="mango-continue-rail__cover">
        {item.cover_url ? (
          <img src={item.cover_url} alt="" loading="lazy" />
        ) : (
          <div className="mango-card__placeholder" />
        )}
      </div>
      <div className="mango-continue-rail__meta">
        <h3 className="mango-continue-rail__title">{item.name}</h3>
        <p className="mango-continue-rail__page">
          {item.page > 0
            ? `${item.page} / ${item.pages} ${t('page')}`
            : `${item.pages} ${t('page')}`}
        </p>
        <ProgressBar value={item.progress} />
        <span className="mango-continue-rail__btn mango-btn mango-btn--primary">
          <Icon icon={icons.continue} size={16} />
          {t('continue')}
        </span>
      </div>
    </AppLink>
  );
}
