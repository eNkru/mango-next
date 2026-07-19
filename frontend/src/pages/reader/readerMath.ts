import type { ReaderDimension } from './types';

/** Clamp one-based page into [1, pages]. Empty book returns 1. */
export function clampPage(page: number, pages: number): number {
  if (!Number.isFinite(page) || pages < 1) return 1;
  return Math.min(Math.max(Math.trunc(page), 1), pages);
}

/**
 * Convert one-based public URL page to the index used by GET /api/page.
 * Runtime page endpoint is 1-based (legacy reader.js), matching archive ReadPage.
 */
export function urlPageToApiIndex(page: number): number {
  return Math.max(1, Math.trunc(page));
}

/** Convert API/page list index (1-based) to public URL page. */
export function apiIndexToUrlPage(index: number): number {
  return Math.max(1, Math.trunc(index));
}

/**
 * Whether left key / left click zone should go next.
 * RTL inverts LTR navigation.
 */
export function nextDirectionIsLeft(enableRightToLeft: boolean): boolean {
  return enableRightToLeft;
}

/**
 * Legacy progress throttle:
 * save when first/last, distance >= 5, or long-page title (avg h/w > 2).
 */
export function shouldSaveProgress(
  page: number,
  lastSavedPage: number,
  pages: number,
  longPages: boolean,
): boolean {
  const p = Math.trunc(page);
  if (pages < 1) return false;
  if (p === 1 || p === pages) return true;
  if (longPages) return true;
  return Math.abs(p - lastSavedPage) >= 5;
}

export function isLongPageTitle(dimensions: ReaderDimension[]): boolean {
  if (!dimensions.length) return false;
  const avg =
    dimensions.reduce((acc, d) => {
      const w = d.width > 0 ? d.width : 1;
      return acc + d.height / w;
    }, 0) / dimensions.length;
  return avg > 2;
}

export function readerPageImagePath(tid: string, eid: string, page: number): string {
  const idx = urlPageToApiIndex(page);
  return `api/page/${encodeURIComponent(tid)}/${encodeURIComponent(eid)}/${idx}`;
}
