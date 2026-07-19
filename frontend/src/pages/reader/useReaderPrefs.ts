import { useCallback, useState } from 'react';
import type { ReaderFitType, ReaderMode, ReaderPrefs } from './types';

const KEYS = {
  mode: 'mango.reader.mode',
  margin: 'mango.reader.margin',
  fitType: 'mango.reader.fitType',
  preloadLookahead: 'mango.reader.preloadLookahead',
  enableFlipAnimation: 'mango.reader.enableFlipAnimation',
  enableRightToLeft: 'mango.reader.enableRightToLeft',
} as const;

const DEFAULTS: ReaderPrefs = {
  mode: 'continuous',
  margin: 30,
  fitType: 'vert',
  preloadLookahead: 3,
  enableFlipAnimation: true,
  enableRightToLeft: false,
};

function readMode(): ReaderMode {
  const v = localStorage.getItem(KEYS.mode);
  return v === 'paged' ? 'paged' : 'continuous';
}

function readFitType(): ReaderFitType {
  const v = localStorage.getItem(KEYS.fitType);
  if (v === 'horz' || v === 'original' || v === 'vert') return v;
  return DEFAULTS.fitType;
}

function readNumber(key: string, fallback: number, min: number, max: number): number {
  const raw = localStorage.getItem(key);
  if (raw == null) return fallback;
  const n = Number(raw);
  if (!Number.isFinite(n)) return fallback;
  return Math.min(max, Math.max(min, n));
}

function readBool(key: string, fallback: boolean): boolean {
  const raw = localStorage.getItem(key);
  if (raw == null) return fallback;
  return raw === 'true';
}

function loadPrefs(): ReaderPrefs {
  return {
    mode: readMode(),
    margin: readNumber(KEYS.margin, DEFAULTS.margin, 0, 120),
    fitType: readFitType(),
    preloadLookahead: readNumber(KEYS.preloadLookahead, DEFAULTS.preloadLookahead, 0, 5),
    enableFlipAnimation: readBool(KEYS.enableFlipAnimation, DEFAULTS.enableFlipAnimation),
    enableRightToLeft: readBool(KEYS.enableRightToLeft, DEFAULTS.enableRightToLeft),
  };
}

export function useReaderPrefs() {
  const [prefs, setPrefsState] = useState<ReaderPrefs>(loadPrefs);

  const setPrefs = useCallback((patch: Partial<ReaderPrefs>) => {
    setPrefsState((prev) => {
      const next = { ...prev, ...patch };
      localStorage.setItem(KEYS.mode, next.mode);
      localStorage.setItem(KEYS.margin, String(next.margin));
      localStorage.setItem(KEYS.fitType, next.fitType);
      localStorage.setItem(KEYS.preloadLookahead, String(next.preloadLookahead));
      localStorage.setItem(KEYS.enableFlipAnimation, String(next.enableFlipAnimation));
      localStorage.setItem(KEYS.enableRightToLeft, String(next.enableRightToLeft));
      return next;
    });
  }, []);

  return { prefs, setPrefs };
}
