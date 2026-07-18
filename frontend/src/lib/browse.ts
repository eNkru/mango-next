export type BrowseTitle = {
  id: string; name: string; display_name: string; file_name: string; sort_name: string; cover_url: string;
  entry_count: number; progress: number; modified_at: number; hidden: boolean; tags?: string[];
};

export type BrowseEntry = {
  id: string; title_id: string; name: string; file_name: string; sort_name: string;
  cover_url: string; pages: number; page: number; progress: number; modified_at: number;
};

export type SortMode = 'natural' | 'title' | 'modified' | 'progress';

const collator = new Intl.Collator(undefined, { numeric: true, sensitivity: 'base' });

export function sortBrowseItems<T extends { name: string; file_name: string; sort_name: string; modified_at: number; progress: number }>(
  items: T[], mode: SortMode, ascending: boolean,
): T[] {
  return [...items].sort((left, right) => {
    let value = 0;
    if (mode === 'modified') value = left.modified_at - right.modified_at;
    else if (mode === 'progress') value = left.progress - right.progress;
    else if (mode === 'title') value = collator.compare(left.name, right.name);
    else value = collator.compare(left.sort_name || left.file_name, right.sort_name || right.file_name);
    return ascending ? value : -value;
  });
}

export function filterBrowseItems<T extends { name: string; file_name: string }>(items: T[], query: string): T[] {
  const value = query.trim().toLocaleLowerCase();
  if (!value) return items;
  return items.filter((item) => `${item.name} ${item.file_name}`.toLocaleLowerCase().includes(value));
}
