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
    {data?.continue_reading.length ? <section className="mango-browse-section"><h2>{t('continueReading')}</h2><div className="mango-continue-grid">{data.continue_reading.map((item) => <article className="mango-continue-card" key={item.id}>
      <img src={item.cover_url} alt="" /><div><h3>{item.name}</h3><p>{item.page > 0 ? `${item.page} / ${item.pages} ${t('page')}` : `${item.pages} ${t('page')}`}</p><ProgressBar value={item.progress} /><div className="mango-actions"><a className="mango-btn mango-btn--primary" href={baseUrl(`reader/${encodeURIComponent(item.title_id)}/${encodeURIComponent(item.id)}`)}>{t('continue')}</a><a className="mango-btn" href={baseUrl(`book/${encodeURIComponent(item.title_id)}`)}>{t('open')}</a></div></div>
    </article>)}</div></section> : null}
    {data?.start_reading.length ? <Rail title={t('startReading')} items={data.start_reading} /> : null}
    {data?.recently_added.length ? <Rail title={t('recentlyAdded')} items={data.recently_added} /> : null}
  </AppShell>;
}

function Rail({ title, items }: { title: string; items: BrowseTitle[] }) {
  return <section className="mango-browse-section"><h2>{title}</h2><div className="mango-poster-rail">{items.map((item) => <PosterCard key={item.id} item={item} />)}</div></section>;
}
