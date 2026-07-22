import { create } from 'zustand';
import type { ReaderFitType, ReaderMode, ReaderPrefs } from '../pages/reader/types';

export const READER_PREF_KEYS = {
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
  const v = localStorage.getItem(READER_PREF_KEYS.mode);
  return v === 'paged' ? 'paged' : 'continuous';
}

function readFitType(): ReaderFitType {
  const v = localStorage.getItem(READER_PREF_KEYS.fitType);
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

export function loadReaderPrefs(): ReaderPrefs {
  return {
    mode: readMode(),
    margin: readNumber(READER_PREF_KEYS.margin, DEFAULTS.margin, 0, 120),
    fitType: readFitType(),
    preloadLookahead: readNumber(
      READER_PREF_KEYS.preloadLookahead,
      DEFAULTS.preloadLookahead,
      0,
      5,
    ),
    enableFlipAnimation: readBool(
      READER_PREF_KEYS.enableFlipAnimation,
      DEFAULTS.enableFlipAnimation,
    ),
    enableRightToLeft: readBool(READER_PREF_KEYS.enableRightToLeft, DEFAULTS.enableRightToLeft),
  };
}

function writePrefs(prefs: ReaderPrefs) {
  localStorage.setItem(READER_PREF_KEYS.mode, prefs.mode);
  localStorage.setItem(READER_PREF_KEYS.margin, String(prefs.margin));
  localStorage.setItem(READER_PREF_KEYS.fitType, prefs.fitType);
  localStorage.setItem(READER_PREF_KEYS.preloadLookahead, String(prefs.preloadLookahead));
  localStorage.setItem(READER_PREF_KEYS.enableFlipAnimation, String(prefs.enableFlipAnimation));
  localStorage.setItem(READER_PREF_KEYS.enableRightToLeft, String(prefs.enableRightToLeft));
}

type ReaderPrefsState = {
  prefs: ReaderPrefs;
  setPrefs: (patch: Partial<ReaderPrefs>) => void;
  rehydrateFromStorage: () => void;
};

export const useReaderPrefsStore = create<ReaderPrefsState>((set, get) => ({
  prefs: loadReaderPrefs(),

  setPrefs: (patch) => {
    const next = { ...get().prefs, ...patch };
    writePrefs(next);
    set({ prefs: next });
  },

  rehydrateFromStorage: () => {
    set({ prefs: loadReaderPrefs() });
  },
}));

export function isReaderPrefKey(key: string | null): boolean {
  if (!key) return false;
  return key.startsWith('mango.reader.');
}
