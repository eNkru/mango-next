import { isReaderPrefKey, useReaderPrefsStore } from './readerPrefsStore';
import { THEME_STORAGE_KEYS, useThemeStore } from './themeStore';

let started = false;

/** Multi-tab sync: other tabs' localStorage writes fire `storage` on this window. */
export function startPrefsStorageSync(): () => void {
  if (started || typeof window === 'undefined') return () => {};
  started = true;

  const onStorage = (event: StorageEvent) => {
    if (!event.key) {
      useThemeStore.getState().rehydrateFromStorage();
      useReaderPrefsStore.getState().rehydrateFromStorage();
      return;
    }
    if ((THEME_STORAGE_KEYS as readonly string[]).includes(event.key)) {
      useThemeStore.getState().rehydrateFromStorage();
    }
    if (isReaderPrefKey(event.key)) {
      useReaderPrefsStore.getState().rehydrateFromStorage();
    }
  };

  window.addEventListener('storage', onStorage);
  return () => {
    window.removeEventListener('storage', onStorage);
    started = false;
  };
}
