import { useCallback, useEffect, useState } from 'react';
import { apiFetch } from '../lib/api';
import type { BrowseEntry, BrowseTitle } from '../lib/browse';
import { baseUrl } from '../lib/baseUrl';
import { useI18n } from '../lib/i18n';
import { PosterCard, ProgressBar } from '../browse/BrowseComponents';
import { AppShell } from '../shell/AppShell';
import { pushAlert } from '../shell/AlertHost';
import { ErrorState, LoadingState } from '../shell/StatePanels';

type HomeResponse = {
  new_user: boolean; empty_library: boolean; is_admin: boolean; library_path: string;
  continue_reading: BrowseEntry[]; start_reading: BrowseTitle[]; recently_added: BrowseTitle[];
};

const LIST_PREVIEW = 3;

function readerUrl(item: BrowseEntry) {
  return baseUrl(`reader/${encodeURIComponent(item.title_id)}/${encodeURIComponent(item.id)}`);
}

export function HomePage() {
  const { t } = useI18n();
  const [data, setData] = useState<HomeResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const load = useCallback(async () => {
    setError(null);
    try { setData(await apiFetch<HomeResponse>('api/home')); }
    catch (err) { const message = err instanceof Error ? err.message : 'Failed to load'; setError(message); pushAlert(message, 'danger'); }
  }, []);
  useEffect(() => { void load(); }, [load]);

  return <AppShell title={t('home')} subtitle={t('homeSubtitle')}>
    {!data && !error ? <LoadingState message={t('loading')} /> : null}
    {error ? <><ErrorState message={error} /><button className="mango-btn" type="button" onClick={() => void load()}>{t('retry')}</button></> : null}
    {data?.empty_library ? <section className="mango-empty-hero"><h2>{t('emptyLibrary')}</h2><p>{data.is_admin ? t('emptyLibraryAdmin') : t('emptyLibraryUser')}</p>{data.is_admin ? <code>{data.library_path}</code> : null}</section> : null}
    {data && !data.empty_library && data.new_user ? <section className="mango-welcome"><h2>{t('welcome')}</h2><p>{t('welcomeBody')}</p></section> : null}
    {data?.continue_reading.length ? <ContinueSection items={data.continue_reading} /> : null}
    {data?.start_reading.length ? <Rail title={t('startReading')} items={data.start_reading} /> : null}
    {data?.recently_added.length ? <Rail title={t('recentlyAdded')} items={data.recently_added} /> : null}
  </AppShell>;
}

function ContinueSection({ items }: { items: BrowseEntry[] }) {
  const { t } = useI18n();
  const [expanded, setExpanded] = useState(false);
  const primary = items[0];
  const rest = items.slice(1);
  const visible = expanded ? rest : rest.slice(0, LIST_PREVIEW);
  return <section className="mango-browse-section"><h2>{t('continueReading')}</h2><div className="mango-continue">
    <article className="mango-continue-hero">
      <img src={primary.cover_url} alt="" />
      <div>
        <h3>{primary.name}</h3>
        <p>{primary.page > 0 ? `${primary.page} / ${primary.pages} ${t('page')}` : `${primary.pages} ${t('page')}`}</p>
        <ProgressBar value={primary.progress} />
        <div className="mango-actions"><a className="mango-btn mango-btn--primary" href={readerUrl(primary)}>{t('continue')}</a></div>
      </div>
    </article>
    {rest.length ? <ul className="mango-continue-list">{visible.map((item) => <li key={item.id}><a className="mango-continue-row" href={readerUrl(item)}><img src={item.cover_url} alt="" /><div><h3>{item.name}</h3><ProgressBar value={item.progress} /></div></a></li>)}</ul> : null}
    {rest.length > LIST_PREVIEW ? <button className="mango-continue-more" type="button" aria-expanded={expanded} onClick={() => setExpanded((value) => !value)}>{expanded ? t('showLess') : t('showMore')}</button> : null}
  </div></section>;
}

function Rail({ title, items }: { title: string; items: BrowseTitle[] }) {
  return <section className="mango-browse-section"><h2>{title}</h2><div className="mango-poster-rail">{items.map((item) => <PosterCard key={item.id} item={item} />)}</div></section>;
}
