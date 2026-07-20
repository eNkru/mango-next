import { useI18n } from '../../lib/i18n';
import { baseUrl } from '../../lib/baseUrl';
import { LanguageSelect } from '../../shell/LanguageSelect';

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
  const { t } = useI18n();
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
        <LanguageSelect />
        <a className="mango-btn mango-btn--primary" href={exitUrl || baseUrl()}>
          {t('exitReader')}
        </a>
      </div>
    </header>
  );
}
