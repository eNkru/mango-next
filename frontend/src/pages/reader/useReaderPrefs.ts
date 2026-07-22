import { useReaderPrefsStore } from '../../lib/readerPrefsStore';

/** Thin hook over reader prefs Zustand store (same API as before). */
export function useReaderPrefs() {
  const prefs = useReaderPrefsStore((s) => s.prefs);
  const setPrefs = useReaderPrefsStore((s) => s.setPrefs);
  return { prefs, setPrefs };
}
