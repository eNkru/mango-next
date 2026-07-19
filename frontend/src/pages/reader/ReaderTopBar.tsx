import type { Language } from '../../lib/i18n';
import { useI18n } from '../../lib/i18n';
import { baseUrl } from '../../lib/baseUrl';

type Props = {
  visible: boolean;
  title: string;
  entryName: string;
  page: number;
  pages: number;
  exitUrl: string;
  onOpenControls: () => void;
  onPointerEnter: () => void;
  onPointerLeave: () => void;
};

export function ReaderTopBar({
  visible,
  title,
  entryName,
  page,
  pages,
  exitUrl,
  onOpenControls,
  onPointerEnter,
  onPointerLeave,
}: Props) {
  const { language, setLanguage, t } = useI18n();
  const pct = pages > 0 ? ((page / pages) * 100).toFixed(1) : '0.0';

  return (
    <header
      className={`mango-reader-topbar${visible ? ' is-visible' : ''}`}
      onPointerEnter={onPointerEnter}
      onPointerLeave={onPointerLeave}
    >
      <div className="mango-reader-topbar__left">
        <a className="mango-reader-topbar__brand" href={baseUrl()}>
          Mango
        </a>
        <button type="button" className="mango-btn mango-btn--ghost" onClick={onOpenControls}>
          {t('readerControls')}
        </button>
      </div>
      <div className="mango-reader-topbar__center" title={`${title} · ${entryName}`}>
        <span className="mango-reader-topbar__title">{entryName || title}</span>
        <span className="mango-reader-topbar__progress">
          {page}/{pages} ({pct}%)
        </span>
      </div>
      <div className="mango-reader-topbar__right">
        <label className="mango-language">
          <span className="sr-only">{t('language')}</span>
          <select
            value={language}
            onChange={(event) => setLanguage(event.target.value as Language)}
            aria-label={t('language')}
          >
            <option value="zh-cn">简体中文</option>
            <option value="zh-tw">繁體中文</option>
            <option value="en">English</option>
          </select>
        </label>
        <a className="mango-btn mango-btn--primary" href={exitUrl || baseUrl()}>
          {t('exitReader')}
        </a>
      </div>
    </header>
  );
}
