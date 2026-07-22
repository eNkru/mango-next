import { useCallback, useEffect, useState } from 'react';
import { apiFetch } from '../lib/api';
import { AppLink } from '../lib/AppLink';
import { useI18n } from '../lib/i18n';
import { pushAlert } from '../shell/AlertHost';
import { AppShell } from '../shell/AppShell';
import { Icon } from '../shell/Icon';
import { icons } from '../shell/icons';
import { ErrorState } from '../shell/StatePanels';

type ScanProgress = {
  success?: boolean;
  running?: boolean;
  titles?: number;
  milliseconds?: number;
  error?: string;
};

type ThumbProgress = {
  success?: boolean;
  running?: boolean;
  progress?: number;
};

type ActionError = {
  kind: 'scan' | 'thumb';
  message: string;
};

export function AdminPage() {
  const { t } = useI18n();
  const [scanning, setScanning] = useState(false);
  const [scanTitles, setScanTitles] = useState(0);
  const [scanMs, setScanMs] = useState(-1);
  const [generating, setGenerating] = useState(false);
  const [thumbProgress, setThumbProgress] = useState(0);
  const [actionError, setActionError] = useState<ActionError | null>(null);

  const pollScan = useCallback(async () => {
    try {
      const data = await apiFetch<ScanProgress>('api/admin/scan_progress');
      setScanning(Boolean(data.running));
      if (!data.running) {
        setScanTitles(data.titles ?? 0);
        setScanMs(typeof data.milliseconds === 'number' ? data.milliseconds : -1);
        if (data.error) pushAlert(data.error, 'danger');
      }
      return Boolean(data.running);
    } catch (err) {
      setScanning(false);
      pushAlert(err instanceof Error ? err.message : t('scanFailed'), 'danger');
      return false;
    }
  }, [t]);

  const pollThumb = useCallback(async () => {
    try {
      const data = await apiFetch<ThumbProgress>('api/admin/thumbnail_progress');
      setGenerating(Boolean(data.running));
      setThumbProgress(typeof data.progress === 'number' ? data.progress : 0);
      return Boolean(data.running);
    } catch {
      setGenerating(false);
      return false;
    }
  }, []);

  useEffect(() => {
    void pollScan();
    void pollThumb();
  }, [pollScan, pollThumb]);

  useEffect(() => {
    if (!scanning) return;
    const id = window.setInterval(() => {
      void pollScan().then((running) => {
        if (!running) window.clearInterval(id);
      });
    }, 1000);
    return () => window.clearInterval(id);
  }, [scanning, pollScan]);

  useEffect(() => {
    if (!generating) return;
    const id = window.setInterval(() => {
      void pollThumb().then((running) => {
        if (!running) window.clearInterval(id);
      });
    }, 1000);
    return () => window.clearInterval(id);
  }, [generating, pollThumb]);

  const startScan = async () => {
    if (scanning) return;
    setScanning(true);
    setScanMs(-1);
    setActionError(null);
    try {
      await apiFetch('api/admin/scan', { method: 'POST' });
      await pollScan();
    } catch (err) {
      setScanning(false);
      setActionError({
        kind: 'scan',
        message: err instanceof Error ? err.message : t('scanFailed'),
      });
    }
  };

  const startThumbnails = async () => {
    if (generating) return;
    setGenerating(true);
    setThumbProgress(0);
    setActionError(null);
    try {
      await apiFetch('api/admin/generate_thumbnails', { method: 'POST' });
      await pollThumb();
    } catch (err) {
      setGenerating(false);
      setActionError({
        kind: 'thumb',
        message: err instanceof Error ? err.message : t('thumbFailed'),
      });
    }
  };

  return (
    <AppShell title={t('admin')} subtitle={t('adminSubtitle')}>
      <div className="mango-admin-grid">
        <AppLink className="mango-admin-card" to="admin/user">
          <strong className="mango-admin-card__title">
            <Icon icon={icons.users} size={18} />
            {t('userManagement')}
          </strong>
          <span>{t('userManagementDesc')}</span>
        </AppLink>
        <AppLink className="mango-admin-card" to="admin/missing">
          <strong className="mango-admin-card__title">
            <Icon icon={icons.missing} size={18} />
            {t('missingEntries')}
          </strong>
          <span>{t('missingEntriesDesc')}</span>
        </AppLink>
        <button
          type="button"
          className="mango-admin-card mango-admin-card--action"
          disabled={scanning}
          onClick={() => void startScan()}
        >
          <strong className="mango-admin-card__title">
            <Icon icon={icons.scan} size={18} />
            {t('scanLibrary')}
          </strong>
          <span>
            {scanning
              ? t('scanning')
              : scanMs >= 0
                ? t('scanResult', { count: scanTitles, ms: scanMs })
                : t('scanLibraryDesc')}
          </span>
        </button>
        <button
          type="button"
          className="mango-admin-card mango-admin-card--action"
          disabled={generating}
          onClick={() => void startThumbnails()}
        >
          <strong className="mango-admin-card__title">
            <Icon icon={icons.refresh} size={18} />
            {t('generateThumbnails')}
          </strong>
          <span>
            {generating
              ? `${(thumbProgress * 100).toFixed(1)}%`
              : t('generateThumbnailsDesc')}
          </span>
        </button>
      </div>
      {actionError ? (
        <ErrorState
          message={actionError.message}
          onRetry={() =>
            void (actionError.kind === 'scan' ? startScan() : startThumbnails())
          }
          retryLabel={t('retry')}
        />
      ) : null}
    </AppShell>
  );
}
