import { useI18n } from '../../lib/i18n';
import { AppLink } from '../../lib/AppLink';
import { baseUrl } from '../../lib/baseUrl';
import { Icon } from '../../shell/Icon';
import { icons } from '../../shell/icons';
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
        <AppLink className="mango-reader-topbar__brand" to="">
          <img
            className="mango-topbar__mark"
            src={baseUrl('img/icons/mango-mark.svg')}
            alt=""
            width={24}
            height={24}
          />
          <span className="mango-topbar__wordmark">Mango</span>
        </AppLink>
        <button type="button" className="mango-btn mango-btn--ghost" onClick={onOpenControls}>
          <Icon icon={icons.readerControls} size={16} />
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
        <AppLink className="mango-btn mango-btn--primary" to={exitUrl || ''}>
          <Icon icon={icons.exit} size={16} />
          {t('exitReader')}
        </AppLink>
      </div>
    </header>
  );
}
