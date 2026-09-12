import type { BrowseEntry } from '../lib/browse';

/**
 * Build the reader deep-link for a continue-reading entry.
 * Storage page is 0-based; reader route `:page` is 1-based.
 * When `page <= 0`, omit the page segment (reader defaults to page 1).
 */
export function continueReaderPath(item: BrowseEntry): string {
  const base = `reader/${encodeURIComponent(item.title_id)}/${encodeURIComponent(item.id)}`;
  if (item.page > 0) {
    return `${base}/${item.page + 1}`;
  }
  return base;
}
