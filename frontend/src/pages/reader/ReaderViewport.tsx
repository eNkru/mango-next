import { useEffect, useRef, type CSSProperties } from 'react';
import { baseUrl } from '../../lib/baseUrl';
import { readerPageImagePath } from './readerMath';
import type { ReaderDimension, ReaderFitType, ReaderMode } from './types';

type Props = {
  tid: string;
  eid: string;
  pages: number;
  page: number;
  dimensions: ReaderDimension[];
  mode: ReaderMode;
  margin: number;
  fitType: ReaderFitType;
  enableFlipAnimation: boolean;
  flipSide: 'left' | 'right' | null;
  onImageClick: (page: number) => void;
  onZoneClick: (zoneIsRight: boolean) => void;
  onVisiblePage: (page: number) => void;
};

export function ReaderViewport({
  tid,
  eid,
  pages,
  page,
  dimensions,
  mode,
  margin,
  fitType,
  enableFlipAnimation,
  flipSide,
  onImageClick,
  onZoneClick,
  onVisiblePage,
}: Props) {
  const stripRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (mode !== 'continuous' || !stripRef.current) return;
    const root = stripRef.current;
    const observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (!entry.isIntersecting) continue;
          const id = Number((entry.target as HTMLElement).dataset.page);
          if (Number.isFinite(id)) onVisiblePage(id);
        }
      },
      { root: null, threshold: 0.55 },
    );
    root.querySelectorAll('[data-page]').forEach((el) => observer.observe(el));
    return () => observer.disconnect();
  }, [mode, pages, onVisiblePage]);

  useEffect(() => {
    if (mode !== 'continuous') return;
    const el = document.getElementById(`reader-page-${page}`);
    el?.scrollIntoView({ block: 'start' });
  }, [mode]); // only on mode enter; avoid fighting scroll

  if (mode === 'continuous') {
    return (
      <div className="mango-reader-strip" ref={stripRef}>
        {Array.from({ length: pages }, (_, i) => {
          const p = i + 1;
          const dim = dimensions[i];
          const w = dim?.width || undefined;
          const h = dim?.height || undefined;
          return (
            <img
              key={p}
              id={`reader-page-${p}`}
              data-page={p}
              className="mango-reader-strip__img"
              src={baseUrl(readerPageImagePath(tid, eid, p))}
              alt=""
              width={w}
              height={h}
              loading={p <= page + 2 ? 'eager' : 'lazy'}
              style={{ marginTop: margin, marginBottom: margin }}
              onClick={() => onImageClick(p)}
            />
          );
        })}
      </div>
    );
  }

  const dim = dimensions[page - 1];
  const animClass =
    enableFlipAnimation && flipSide
      ? flipSide === 'left'
        ? ' mango-reader-page--flip-left'
        : ' mango-reader-page--flip-right'
      : '';

  const fitStyle: CSSProperties =
    fitType === 'vert'
      ? { height: '100vh', width: 'auto', maxHeight: '100%', objectFit: 'contain' }
      : fitType === 'horz'
        ? { width: '100vw', height: 'auto', maxWidth: '100%', objectFit: 'contain' }
        : { maxWidth: 'none', maxHeight: 'none', objectFit: 'contain' };

  return (
    <div className="mango-reader-paged">
      <img
        className={`mango-reader-paged__img${animClass}`}
        src={baseUrl(readerPageImagePath(tid, eid, page))}
        alt=""
        width={dim?.width || undefined}
        height={dim?.height || undefined}
        style={fitStyle}
        onClick={() => onImageClick(page)}
      />
      <button
        type="button"
        className="mango-reader-zone mango-reader-zone--left"
        aria-label="prev"
        onClick={() => onZoneClick(false)}
      />
      <button
        type="button"
        className="mango-reader-zone mango-reader-zone--right"
        aria-label="next"
        onClick={() => onZoneClick(true)}
      />
    </div>
  );
}
