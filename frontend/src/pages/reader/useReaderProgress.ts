import { useCallback, useEffect, useRef } from 'react';
import { apiFetch } from '../../lib/api';
import { shouldSaveProgress } from './readerMath';

type Options = {
  tid: string;
  eid: string;
  pages: number;
  longPages: boolean;
  initialSaved?: number;
};

export function useReaderProgress({ tid, eid, pages, longPages, initialSaved = 0 }: Options) {
  const lastSaved = useRef(initialSaved || 0);

  useEffect(() => {
    lastSaved.current = initialSaved || 0;
  }, [tid, eid, initialSaved]);

  const save = useCallback(
    async (page: number, force = false) => {
      const p = Math.trunc(page);
      if (!force && !shouldSaveProgress(p, lastSaved.current, pages, longPages)) {
        return;
      }
      lastSaved.current = p;
      try {
        await apiFetch(
          `api/progress/${encodeURIComponent(tid)}/${p}?eid=${encodeURIComponent(eid)}`,
          { method: 'PUT' },
        );
      } catch {
        // Non-blocking: progress save failures should not break reading.
      }
    },
    [tid, eid, pages, longPages],
  );

  const complete = useCallback(async () => {
    await save(pages, true);
  }, [save, pages]);

  return { save, complete, lastSaved };
}
