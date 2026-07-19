import type { FormEvent } from 'react';
import { useI18n } from '../../lib/i18n';
import type { ReaderEntry, ReaderFitType, ReaderMode, ReaderPrefs } from './types';

type Props = {
  open: boolean;
  entryName: string;
  entryId: string;
  page: number;
  pages: number;
  entries: ReaderEntry[];
  prefs: ReaderPrefs;
  nextEntryUrl: string;
  previousEntryUrl: string;
  exitUrl: string;
  onClose: () => void;
  onJumpPage: (page: number) => void;
  onPrefs: (patch: Partial<ReaderPrefs>) => void;
  onJumpEntry: (eid: string) => void;
  onNextEntry: () => void;
  onExit: () => void;
};

export function ReaderControls({
  open,
  entryName,
  entryId,
  page,
  pages,
  entries,
  prefs,
  nextEntryUrl,
  previousEntryUrl,
  exitUrl,
  onClose,
  onJumpPage,
  onPrefs,
  onJumpEntry,
  onNextEntry,
  onExit,
}: Props) {
  const { t } = useI18n();
  if (!open) return null;

  const pct = pages > 0 ? ((page / pages) * 100).toFixed(1) : '0.0';

  const submitJump = (event: FormEvent) => {
    event.preventDefault();
  };

  return (
    <div className="mango-modal-backdrop mango-reader-modal-backdrop" role="presentation" onClick={onClose}>
      <div
        className="mango-modal mango-reader-controls"
        role="dialog"
        aria-modal="true"
        aria-label={t('readerControls')}
        onClick={(event) => event.stopPropagation()}
      >
        <header className="mango-reader-controls__header">
          <div>
            <h2>{entryName}</h2>
            <p className="mango-file-name">{entryId}</p>
          </div>
          <button type="button" className="mango-btn" onClick={onClose} aria-label={t('cancel')}>
            ×
          </button>
        </header>

        <p className="mango-reader-controls__progress">
          {t('progress')}: {page}/{pages} ({pct}%)
        </p>

        <form className="mango-field" onSubmit={submitJump}>
          <label>
            <span>{t('jumpToPage')}</span>
            <select
              className="mango-input"
              value={page}
              onChange={(event) => onJumpPage(Number(event.target.value))}
            >
              {Array.from({ length: pages }, (_, i) => {
                const p = i + 1;
                return (
                  <option key={p} value={p}>
                    {p}
                  </option>
                );
              })}
            </select>
          </label>
        </form>

        <label className="mango-field">
          <span>{t('readingMode')}</span>
          <select
            className="mango-input"
            value={prefs.mode}
            onChange={(event) => onPrefs({ mode: event.target.value as ReaderMode })}
          >
            <option value="continuous">{t('modeContinuous')}</option>
            <option value="paged">{t('modePaged')}</option>
          </select>
        </label>

        {prefs.mode === 'paged' ? (
          <label className="mango-field">
            <span>{t('fitPage')}</span>
            <select
              className="mango-input"
              value={prefs.fitType}
              onChange={(event) => onPrefs({ fitType: event.target.value as ReaderFitType })}
            >
              <option value="vert">{t('fitHeight')}</option>
              <option value="horz">{t('fitWidth')}</option>
              <option value="original">{t('fitOriginal')}</option>
            </select>
          </label>
        ) : (
          <label className="mango-field">
            <span>
              {t('pageMargin')}: {prefs.margin}px
            </span>
            <input
              className="mango-input"
              type="range"
              min={0}
              max={80}
              value={prefs.margin}
              onChange={(event) => onPrefs({ margin: Number(event.target.value) })}
            />
          </label>
        )}

        <label className="mango-field">
          <span>
            {t('preloadLookahead')}: {prefs.preloadLookahead}
          </span>
          <input
            className="mango-input"
            type="range"
            min={0}
            max={5}
            value={prefs.preloadLookahead}
            onChange={(event) => onPrefs({ preloadLookahead: Number(event.target.value) })}
          />
        </label>

        <label className="mango-field mango-field--row">
          <input
            type="checkbox"
            checked={prefs.enableFlipAnimation}
            onChange={(event) => onPrefs({ enableFlipAnimation: event.target.checked })}
          />
          <span>{t('flipAnimation')}</span>
        </label>

        <label className="mango-field mango-field--row">
          <input
            type="checkbox"
            checked={prefs.enableRightToLeft}
            onChange={(event) => onPrefs({ enableRightToLeft: event.target.checked })}
          />
          <span>{t('rightToLeft')}</span>
        </label>

        {entries.length > 1 ? (
          <label className="mango-field">
            <span>{t('jumpToEntry')}</span>
            <select
              className="mango-input"
              value={entryId}
              onChange={(event) => onJumpEntry(event.target.value)}
            >
              {entries.map((entry) => (
                <option key={entry.id} value={entry.id}>
                  {entry.name}
                </option>
              ))}
            </select>
          </label>
        ) : null}

        <div className="mango-actions mango-reader-controls__actions">
          {previousEntryUrl ? (
            <a className="mango-btn" href={previousEntryUrl}>
              {t('previousEntry')}
            </a>
          ) : null}
          {nextEntryUrl ? (
            <button type="button" className="mango-btn mango-btn--primary" onClick={onNextEntry}>
              {t('nextEntry')}
            </button>
          ) : (
            <button type="button" className="mango-btn mango-btn--primary" onClick={onExit}>
              {t('exitReader')}
            </button>
          )}
          {exitUrl ? (
            <button type="button" className="mango-btn" onClick={onExit}>
              {t('exitReader')}
            </button>
          ) : null}
        </div>
      </div>
    </div>
  );
}
