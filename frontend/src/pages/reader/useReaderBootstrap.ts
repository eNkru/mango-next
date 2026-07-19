import { useCallback, useEffect, useState } from 'react';
import { apiFetch } from '../../lib/api';
import type { ReaderBootstrap } from './types';

type State =
  | { status: 'loading' }
  | { status: 'error'; message: string }
  | { status: 'ready'; data: ReaderBootstrap };

export function useReaderBootstrap(tid: string, eid: string) {
  const [state, setState] = useState<State>({ status: 'loading' });

  const load = useCallback(async () => {
    if (!tid || !eid) {
      setState({ status: 'error', message: 'Missing title or entry id' });
      return;
    }
    setState({ status: 'loading' });
    try {
      const res = await apiFetch<{ success: boolean; data: ReaderBootstrap }>(
        `api/reader/${encodeURIComponent(tid)}/${encodeURIComponent(eid)}`,
      );
      if (!res.data || !res.data.entry || res.data.entry.pages <= 0) {
        setState({ status: 'error', message: 'Entry has no pages' });
        return;
      }
      setState({ status: 'ready', data: res.data });
    } catch (err) {
      setState({
        status: 'error',
        message: err instanceof Error ? err.message : 'Failed to load reader',
      });
    }
  }, [tid, eid]);

  useEffect(() => {
    void load();
  }, [load]);

  return { state, reload: load };
}
