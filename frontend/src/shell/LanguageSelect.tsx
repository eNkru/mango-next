import { useI18n, type Language } from '../lib/i18n';

type Props = {
  className?: string;
};

/** Shared language control; persists via useI18n → localStorage mango-language. */
export function LanguageSelect({ className = 'mango-language' }: Props) {
  const { language, setLanguage, t } = useI18n();
  return (
    <label className={className}>
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
  );
}
