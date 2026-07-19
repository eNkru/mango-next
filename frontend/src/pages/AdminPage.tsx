import { useCallback, useEffect, useState } from 'react';
import { apiFetch } from '../lib/api';
import { baseUrl } from '../lib/baseUrl';
import { useI18n } from '../lib/i18n';
import { pushAlert } from '../shell/AlertHost';
import { AppShell } from '../shell/AppShell';

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

export function AdminPage() {
  const { t } = useI18n();
  const [scanning, setScanning] = useState(false);
  const [scanTitles, setScanTitles] = useState(0);
  const [scanMs, setScanMs] = useState(-1);
  const [generating, setGenerating] = useState(false);
  const [thumbProgress, setThumbProgress] = useState(0);

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
    try {
      await apiFetch('api/admin/scan', { method: 'POST' });
      await pollScan();
    } catch (err) {
      setScanning(false);
      pushAlert(err instanceof Error ? err.message : t('scanFailed'), 'danger');
    }
  };

  const startThumbnails = async () => {
    if (generating) return;
    setGenerating(true);
    setThumbProgress(0);
    try {
      await apiFetch('api/admin/generate_thumbnails', { method: 'POST' });
      await pollThumb();
    } catch (err) {
      setGenerating(false);
      pushAlert(err instanceof Error ? err.message : t('thumbFailed'), 'danger');
    }
  };

  return (
    <AppShell title={t('admin')} subtitle={t('adminSubtitle')}>
      <div className="mango-admin-grid">
        <a className="mango-admin-card" href={baseUrl('admin/user')}>
          <strong>{t('userManagement')}</strong>
          <span>{t('userManagementDesc')}</span>
        </a>
        <a className="mango-admin-card" href={baseUrl('admin/missing')}>
          <strong>{t('missingEntries')}</strong>
          <span>{t('missingEntriesDesc')}</span>
        </a>
        <button
          type="button"
          className="mango-admin-card mango-admin-card--action"
          disabled={scanning}
          onClick={() => void startScan()}
        >
          <strong>{t('scanLibrary')}</strong>
          <span>
            {scanning
              ? t('scanning')
              : scanMs >= 0
                ? t('scanResult').replace('{count}', String(scanTitles)).replace('{ms}', String(scanMs))
                : t('scanLibraryDesc')}
          </span>
        </button>
        <button
          type="button"
          className="mango-admin-card mango-admin-card--action"
          disabled={generating}
          onClick={() => void startThumbnails()}
        >
          <strong>{t('generateThumbnails')}</strong>
          <span>
            {generating
              ? `${(thumbProgress * 100).toFixed(1)}%`
              : t('generateThumbnailsDesc')}
          </span>
        </button>
      </div>
    </AppShell>
  );
}
