import { useCallback, useEffect, useRef, useState } from 'react';
import { baseUrl } from '../../lib/baseUrl';
import { clampPage, nextDirectionIsLeft, readerPageImagePath } from './readerMath';

type Options = {
  tid: string;
  eid: string;
  pages: number;
  initialPage: number;
  enableRightToLeft: boolean;
  preloadLookahead: number;
  mode: 'continuous' | 'paged';
  onPageChange?: (page: number) => void;
};

export function useReaderNavigation({
  tid,
  eid,
  pages,
  initialPage,
  enableRightToLeft,
  preloadLookahead,
  mode,
  onPageChange,
}: Options) {
  const [page, setPageState] = useState(() => clampPage(initialPage, pages));
  const pageRef = useRef(page);
  pageRef.current = page;

  const replaceUrl = useCallback(
    (next: number) => {
      const clamped = clampPage(next, pages);
      const path = baseUrl(`reader/${encodeURIComponent(tid)}/${encodeURIComponent(eid)}/${clamped}`);
      history.replaceState(null, '', path);
    },
    [tid, eid, pages],
  );

  const preload = useCallback(
    (from: number) => {
      const limit = Math.min(from + preloadLookahead, pages);
      for (let p = from + 1; p <= limit; p += 1) {
        const img = new Image();
        img.src = baseUrl(readerPageImagePath(tid, eid, p));
      }
    },
    [tid, eid, pages, preloadLookahead],
  );

  const setPage = useCallback(
    (next: number, opts?: { replace?: boolean }) => {
      const clamped = clampPage(next, pages);
      setPageState(clamped);
      if (opts?.replace !== false) replaceUrl(clamped);
      onPageChange?.(clamped);
      preload(clamped);
    },
    [pages, replaceUrl, onPageChange, preload],
  );

  const flip = useCallback(
    (isNext: boolean): boolean | undefined => {
      if (mode === 'continuous') return undefined;
      const delta = isNext ? 1 : -1;
      const next = pageRef.current + delta;
      if (next < 1) return true;
      if (next > pages) return false;
      setPage(next);
      return true;
    },
    [mode, pages, setPage],
  );

  const flipWithRtl = useCallback(
    (zoneIsRight: boolean) => {
      // zoneIsRight=true means right-hand zone / ArrowRight
      const goNext = zoneIsRight !== nextDirectionIsLeft(enableRightToLeft);
      return flip(goNext);
    },
    [enableRightToLeft, flip],
  );

  useEffect(() => {
    setPageState(clampPage(initialPage, pages));
  }, [initialPage, pages, tid, eid]);

  useEffect(() => {
    preload(page);
  }, [page, preload]);

  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      if (mode === 'continuous') return;
      const target = event.target as HTMLElement | null;
      if (target && (target.tagName === 'INPUT' || target.tagName === 'SELECT' || target.tagName === 'TEXTAREA')) {
        return;
      }
      if (event.key === 'ArrowLeft' || event.key === 'k') {
        event.preventDefault();
        flipWithRtl(false);
      } else if (event.key === 'ArrowRight' || event.key === 'j') {
        event.preventDefault();
        flipWithRtl(true);
      }
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [mode, flipWithRtl]);

  return { page, setPage, flip, flipWithRtl };
}
