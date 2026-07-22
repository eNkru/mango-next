import { LANGUAGE_STORAGE_KEY, useI18nStore } from './i18n';
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
      useI18nStore.getState().rehydrateFromStorage();
      return;
    }
    if ((THEME_STORAGE_KEYS as readonly string[]).includes(event.key)) {
      useThemeStore.getState().rehydrateFromStorage();
    }
    if (isReaderPrefKey(event.key)) {
      useReaderPrefsStore.getState().rehydrateFromStorage();
    }
    if (event.key === LANGUAGE_STORAGE_KEY) {
      useI18nStore.getState().rehydrateFromStorage();
    }
  };

  window.addEventListener('storage', onStorage);
  return () => {
    window.removeEventListener('storage', onStorage);
    started = false;
  };
}
