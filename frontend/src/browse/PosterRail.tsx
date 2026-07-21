import { useCallback, useEffect, useRef, useState } from 'react';
import type { BrowseTitle } from '../lib/browse';
import { useI18n } from '../lib/i18n';
import { PosterCard } from './BrowseComponents';
import { Icon } from '../shell/Icon';
import { icons } from '../shell/icons';

export function PosterRail({ title, items }: { title: string; items: BrowseTitle[] }) {
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
      <h2>{title}</h2>
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
        <div ref={trackRef} className="mango-poster-rail">
          {items.map((item) => (
            <PosterCard key={item.id} item={item} />
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
