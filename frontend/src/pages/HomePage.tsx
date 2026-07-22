import { useCallback, useEffect, useState } from 'react';
import { apiFetch } from '../lib/api';
import type { BrowseEntry, BrowseTitle } from '../lib/browse';
import { useI18n } from '../lib/i18n';
import { ContinueCarousel } from '../browse/ContinueCarousel';
import { PosterRail } from '../browse/PosterRail';
import { AppShell } from '../shell/AppShell';
import { pushAlert } from '../shell/AlertHost';
import { ErrorState } from '../shell/StatePanels';

type HomeResponse = {
  new_user: boolean;
  empty_library: boolean;
  is_admin: boolean;
  library_path: string;
  continue_reading: BrowseEntry[];
  start_reading: BrowseTitle[];
  recently_added: BrowseTitle[];
};

export function HomePage() {
  const { t } = useI18n();
  const [data, setData] = useState<HomeResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const load = useCallback(async () => {
    setError(null);
    try {
      setData(await apiFetch<HomeResponse>('api/home'));
    } catch (err) {
      const message = err instanceof Error ? err.message : t('loadFailed');
      setError(message);
      pushAlert(message, 'danger');
    }
  }, [t]);
  useEffect(() => {
    void load();
  }, [load]);

  const loading = !data && !error;

  return (
    <AppShell title={t('home')} subtitle={t('homeSubtitle')}>
      {error ? (
        <ErrorState message={error} onRetry={() => void load()} retryLabel={t('retry')} />
      ) : null}
      {loading ? (
        <>
          <PosterRail title={t('startReading')} items={[]} loading />
          <PosterRail title={t('recentlyAdded')} items={[]} loading />
        </>
      ) : null}
      {data?.empty_library ? (
        <section className="mango-empty-hero">
          <h2>{t('emptyLibrary')}</h2>
          <p>{data.is_admin ? t('emptyLibraryAdmin') : t('emptyLibraryUser')}</p>
          {data.is_admin ? <code>{data.library_path}</code> : null}
        </section>
      ) : null}
      {data && !data.empty_library && data.new_user ? (
        <section className="mango-welcome">
          <h2>{t('welcome')}</h2>
          <p>{t('welcomeBody')}</p>
        </section>
      ) : null}
      {data?.continue_reading.length ? <ContinueCarousel items={data.continue_reading} /> : null}
      {data?.start_reading.length ? (
        <PosterRail title={t('startReading')} items={data.start_reading} />
      ) : null}
      {data?.recently_added.length ? (
        <PosterRail title={t('recentlyAdded')} items={data.recently_added} />
      ) : null}
    </AppShell>
  );
}
