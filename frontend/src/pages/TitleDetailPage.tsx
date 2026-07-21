import { useCallback, useEffect, useMemo, useState, type FormEvent } from 'react';
import { apiFetch } from '../lib/api';
import { baseUrl } from '../lib/baseUrl';
import { filterBrowseItems, sortBrowseItems, type BrowseEntry, type BrowseTitle, type SortMode } from '../lib/browse';
import { readBoot } from '../lib/boot';
import { useI18n } from '../lib/i18n';
import { BrowseToolbar, PosterCard, ProgressBar } from '../browse/BrowseComponents';
import { AppShell } from '../shell/AppShell';
import { pushAlert } from '../shell/AlertHost';
import { Icon } from '../shell/Icon';
import { icons } from '../shell/icons';
import { EmptyState, ErrorState, LoadingState } from '../shell/StatePanels';

type DetailResponse = { is_admin: boolean; title: BrowseTitle; parents: BrowseTitle[]; tags: string[]; titles: BrowseTitle[]; entries: BrowseEntry[] };
type EditTarget = { kind: 'title'; item: BrowseTitle } | { kind: 'entry'; item: BrowseEntry };
type EntryGroup = { key: string; label: string; items: BrowseEntry[] };

export function TitleDetailPage() {
  const { t } = useI18n(); const tid = readBoot().titleId ?? '';
  const [data, setData] = useState<DetailResponse | null>(null); const [error, setError] = useState<string | null>(null); const [loading, setLoading] = useState(true);
  const [query, setQuery] = useState(''); const [mode, setMode] = useState<SortMode>('natural'); const [ascending, setAscending] = useState(true); const [selected, setSelected] = useState<Set<string>>(new Set());
  const [busy, setBusy] = useState(false); const [editTarget, setEditTarget] = useState<EditTarget | null>(null); const [tag, setTag] = useState('');
  const [collapsedGroups, setCollapsedGroups] = useState<Set<string>>(new Set());
  const load = useCallback(async () => { if (!tid) { setError(t('missingTitleId')); setLoading(false); return; } setLoading(true); setError(null); try { setData(await apiFetch<DetailResponse>(`api/book/${encodeURIComponent(tid)}`)); } catch (err) { const message = err instanceof Error ? err.message : t('loadFailed'); setError(message); pushAlert(message, 'danger'); } finally { setLoading(false); } }, [tid, t]);
  useEffect(() => { void load(); }, [load]);
  const entries = useMemo(() => data ? sortBrowseItems(filterBrowseItems(data.entries, query), mode, ascending) : [], [data, query, mode, ascending]);
  const visibleCount = entries.length;
  const totalCount = data?.entries.length ?? 0;
  const isSearching = query.trim().length > 0;
  const groups = useMemo<EntryGroup[]>(() => {
    if (isSearching) return entries.length ? [{ key: 'all', label: t('chapters'), items: entries }] : [];
    const unread: BrowseEntry[] = []; const reading: BrowseEntry[] = []; const completed: BrowseEntry[] = [];
    for (const entry of entries) {
      if (entry.progress >= 100) completed.push(entry);
      else if (entry.progress > 0) reading.push(entry);
      else unread.push(entry);
    }
    const result: EntryGroup[] = [];
    if (reading.length) result.push({ key: 'reading', label: t('readingGroup'), items: reading });
    if (unread.length) result.push({ key: 'unread', label: t('unreadGroup'), items: unread });
    if (completed.length) result.push({ key: 'completed', label: t('completedGroup'), items: completed });
    return result;
  }, [entries, isSearching, t]);
  const readCount = useMemo(() => data ? data.entries.filter((e) => e.progress >= 100).length : 0, [data]);
  const toggleGroup = (key: string) => setCollapsedGroups((prev) => { const next = new Set(prev); if (next.has(key)) next.delete(key); else next.add(key); return next; });
  const mutate = async (path: string, init: RequestInit, success = t('save')) => { setBusy(true); try { await apiFetch(path, init); pushAlert(success, 'success'); await load(); return true; } catch (err) { pushAlert(err instanceof Error ? err.message : t('actionFailed'), 'danger'); return false; } finally { setBusy(false); } };
  const updateProgress = (entry: BrowseEntry, read: boolean) => mutate(`api/progress/${encodeURIComponent(entry.title_id)}/${read ? entry.pages : 0}?eid=${encodeURIComponent(entry.id)}`, { method: 'PUT' }, read ? t('markRead') : t('markUnread'));
  const bulkProgress = async (action: 'read' | 'unread') => { if (!selected.size) return; if (await mutate(`api/bulk_progress/${action}/${encodeURIComponent(tid)}`, { method: 'PUT', body: JSON.stringify({ ids: [...selected] }) }, action === 'read' ? t('markRead') : t('markUnread'))) setSelected(new Set()); };
  const toggleSelection = (id: string) => setSelected((current) => { const next = new Set(current); if (next.has(id)) next.delete(id); else next.add(id); return next; });
  const addTag = async (event: FormEvent) => { event.preventDefault(); const value = tag.trim(); if (!value) return; if (await mutate(`api/admin/tags/${encodeURIComponent(tid)}/${encodeURIComponent(value)}`, { method: 'PUT' })) setTag(''); };

  const renderEntryCard = (entry: BrowseEntry, isAdmin: boolean) => (
    <article className="mango-entry-card" key={entry.id}>
      {isAdmin ? <input type="checkbox" checked={selected.has(entry.id)} onChange={() => toggleSelection(entry.id)} aria-label={`${t('selected')} ${entry.name}`} /> : null}
      <a className="mango-entry-card__cover" href={baseUrl(`reader/${encodeURIComponent(entry.title_id)}/${encodeURIComponent(entry.id)}`)}>{entry.cover_url ? <img src={entry.cover_url} alt="" loading="lazy" /> : <div className="mango-card__placeholder" />}</a>
      <div className="mango-entry-card__body">
        <h3>{entry.name}</h3>
        <p>{entry.pages} {t('page')}</p>
        <ProgressBar value={entry.progress} />
        <div className="mango-actions">
          <a className="mango-btn mango-btn--primary" href={baseUrl(`reader/${encodeURIComponent(entry.title_id)}/${encodeURIComponent(entry.id)}${entry.page > 0 ? '' : '/1'}`)}><Icon icon={entry.page > 0 ? icons.continue : icons.play} size={16} />{entry.page > 0 ? t('continue') : t('begin')}</a>
          <a className="mango-btn" href={baseUrl(`api/download/${encodeURIComponent(entry.title_id)}/${encodeURIComponent(entry.id)}`)}><Icon icon={icons.download} size={16} />{t('download')}</a>
          <button className="mango-btn" type="button" onClick={() => void updateProgress(entry, entry.progress < 100)}><Icon icon={entry.progress >= 100 ? icons.markUnread : icons.markRead} size={16} />{entry.progress >= 100 ? t('markUnread') : t('markRead')}</button>
          {isAdmin ? <button className="mango-btn" type="button" onClick={() => setEditTarget({ kind: 'entry', item: entry })}><Icon icon={icons.edit} size={16} />{t('edit')}</button> : null}
        </div>
      </div>
    </article>
  );

  return <AppShell title={data?.title.name ?? t('titleDetail')} subtitle={data ? `${data.title.entry_count} ${t('entries')}` : undefined}>
    {loading ? <LoadingState message={t('loading')} /> : null}{error ? <ErrorState message={error} onRetry={() => void load()} retryLabel={t('retry')} /> : null}
    {data ? <>
      <nav className="mango-breadcrumb" aria-label="Breadcrumb"><a href={baseUrl('library')}>{t('library')}</a>{data.parents.map((parent) => <span key={parent.id}>/ <a href={baseUrl(`book/${encodeURIComponent(parent.id)}`)}>{parent.name}</a></span>)}<span>/ {data.title.name}</span></nav>
      <section className="mango-title-overview">
        <div className="mango-title-cover mango-title-cover--hover">{data.title.cover_url ? <img src={data.title.cover_url} alt="" /> : <div className="mango-card__placeholder" />}</div>
        <div>
          <div className="mango-title-heading"><h2>{data.title.name}</h2>{data.title.hidden ? <span className="mango-badge mango-badge--muted">{t('hidden')}</span> : null}</div>
          <p className="mango-reading-summary">{t('readProgress', { read: readCount, total: totalCount })}</p>
          <ProgressBar value={data.title.progress} />
          <p className="mango-file-name mango-file-name--dim">{data.title.file_name}</p>
          <div className="mango-tag-list">{(data.tags ?? []).map((item) => <span className="mango-tag-pill" key={item}><a href={baseUrl(`tags/${encodeURIComponent(item)}`)}>{item}</a>{data.is_admin ? <button type="button" className="mango-btn mango-btn--icon" title={t('remove')} aria-label={t('remove')} onClick={() => void mutate(`api/admin/tags/${encodeURIComponent(tid)}/${encodeURIComponent(item)}`, { method: 'DELETE' })}><Icon icon={icons.close} size={14} /></button> : null}</span>)}</div>
          {data.is_admin ? <><form className="mango-inline-form" onSubmit={(event) => void addTag(event)}><input className="mango-input" value={tag} onChange={(event) => setTag(event.target.value)} placeholder={t('addTag')} /><button className="mango-btn" type="submit" disabled={busy}><Icon icon={icons.add} size={16} />{t('addTag')}</button></form><div className="mango-actions"><button className="mango-btn" type="button" onClick={() => setEditTarget({ kind: 'title', item: data.title })}><Icon icon={icons.edit} size={16} />{t('edit')}</button><button className="mango-btn mango-btn--danger" type="button" onClick={() => void mutate(`api/admin/hidden/${encodeURIComponent(tid)}/${data.title.hidden ? 0 : 1}`, { method: 'PUT' })}><Icon icon={data.title.hidden ? icons.show : icons.hide} size={16} />{data.title.hidden ? t('show') : t('hide')}</button></div></> : null}
        </div>
      </section>
      {data.titles.length ? <section className="mango-browse-section mango-browse-section--divided"><h2>{t('children')}</h2><div className="mango-card-grid">{data.titles.map((item) => <PosterCard key={item.id} item={item} />)}</div></section> : null}
      <section className="mango-browse-section">
        <div className="mango-section-heading">
          <h2>{t('chapters')}{totalCount > 0 ? <span className="mango-entry-count">{t('totalChapters', { count: totalCount })}</span> : null}</h2>
          {selected.size ? <div className="mango-selection-bar"><strong>{selected.size} {t('selected')}</strong><button className="mango-btn" type="button" disabled={busy} onClick={() => void bulkProgress('read')}><Icon icon={icons.markRead} size={16} />{t('markRead')}</button><button className="mango-btn" type="button" disabled={busy} onClick={() => void bulkProgress('unread')}><Icon icon={icons.markUnread} size={16} />{t('markUnread')}</button><button className="mango-btn" type="button" onClick={() => setSelected(new Set())}><Icon icon={icons.close} size={16} />{t('clearSelection')}</button></div> : null}
        </div>
        <BrowseToolbar query={query} onQuery={setQuery} mode={mode} onMode={setMode} ascending={ascending} onAscending={setAscending} extra={<>
           {entries.length > 0 ? <select className="mango-input mango-jump-select" value="" onChange={(e) => { const id = e.target.value; if (id) window.location.href = baseUrl(`reader/${encodeURIComponent(entries[0].title_id)}/${encodeURIComponent(id)}/1`); }} aria-label={t('quickJumpTo')}><option value="">{t('quickJumpTo')}</option>{entries.map((e) => <option key={e.id} value={e.id}>{e.name}</option>)}</select> : null}
          {data.is_admin && entries.length ? <button className="mango-btn" type="button" onClick={() => setSelected(new Set(entries.map((item) => item.id)))}><Icon icon={icons.selectAll} size={16} />{t('selectAll')}</button> : null}
        </>} />
        {!entries.length ? <EmptyState message={t('noResults')} /> : groups.map((group) => {
          const collapsed = collapsedGroups.has(group.key);
          return <div className="mango-entry-group" key={group.key}>
            {!isSearching ? <button className="mango-entry-group__header" type="button" onClick={() => toggleGroup(group.key)} aria-expanded={!collapsed}>
              <span>{group.label} <span className="mango-entry-group__count">{group.items.length}</span></span>
              <Icon icon={collapsed ? icons.chevronDown : icons.chevronUp} size={16} />
            </button> : null}
            {!collapsed ? <div className="mango-entry-grid">{group.items.map((e) => renderEntryCard(e, data.is_admin))}</div> : null}
          </div>;
        })}
        {isSearching ? null : entries.length > 0 ? <p className="mango-entry-summary">{t('totalChapters', { count: visibleCount })}</p> : null}
      </section>
      <EditDialog target={editTarget} tid={tid} busy={busy} onClose={() => setEditTarget(null)} onMutate={mutate} />
    </> : null}
  </AppShell>;
}

function EditDialog({ target, tid, busy, onClose, onMutate }: { target: EditTarget | null; tid: string; busy: boolean; onClose: () => void; onMutate: (path: string, init: RequestInit, success?: string) => Promise<boolean> }) {
  const { t } = useI18n(); const [name, setName] = useState(''); const [sortName, setSortName] = useState(''); const [cover, setCover] = useState<File | null>(null);
  useEffect(() => { setName(target?.item.name ?? ''); setSortName(target?.item.sort_name ?? ''); setCover(null); }, [target]);
  if (!target) return null;
  const eid = target.kind === 'entry' ? target.item.id : undefined;
  const submit = async (event: FormEvent) => { event.preventDefault(); const displayOk = await onMutate(`api/admin/display_name/${encodeURIComponent(tid)}`, { method: 'PUT', body: JSON.stringify({ name, eid }) }); if (!displayOk) return; const sortOk = await onMutate(`api/admin/sort_title/${encodeURIComponent(tid)}`, { method: 'PUT', body: JSON.stringify({ sort_name: sortName, eid }) }); if (!sortOk) return; if (cover) { const form = new FormData(); form.set('file', cover); const query = new URLSearchParams({ tid }); if (eid) query.set('eid', eid); if (!await onMutate(`api/admin/upload/cover?${query}`, { method: 'POST', body: form })) return; } onClose(); };
  return <div className="mango-modal-backdrop" role="presentation" onClick={onClose}><form className="mango-modal mango-edit-form" role="dialog" aria-modal="true" onSubmit={(event) => void submit(event)} onClick={(event) => event.stopPropagation()}><h2>{t('edit')}</h2><label className="mango-field"><span>{t('fileName')}</span><input className="mango-input" value={target.item.file_name} disabled /></label><label className="mango-field"><span>{t('displayName')}</span><input className="mango-input" value={name} required onChange={(event) => setName(event.target.value)} /></label><label className="mango-field"><span>{t('sortName')}</span><input className="mango-input" value={sortName} onChange={(event) => setSortName(event.target.value)} /></label><label className="mango-field"><span>{t('upload')}</span><input className="mango-input" type="file" accept="image/jpeg,image/png" onChange={(event) => setCover(event.target.files?.[0] ?? null)} /></label><div className="mango-modal__actions"><button className="mango-btn" type="button" onClick={onClose}><Icon icon={icons.close} size={16} />{t('cancel')}</button><button className="mango-btn mango-btn--primary" type="submit" disabled={busy}><Icon icon={icons.save} size={16} />{t('save')}</button></div></form></div>;
}
