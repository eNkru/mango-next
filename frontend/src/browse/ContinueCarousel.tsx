import { useCallback, useState, type CSSProperties, type KeyboardEvent } from 'react';
import type { BrowseEntry } from '../lib/browse';
import { baseUrl } from '../lib/baseUrl';
import { useI18n } from '../lib/i18n';
import { ProgressBar } from './BrowseComponents';
import { Icon } from '../shell/Icon';
import { icons } from '../shell/icons';

const MAX_STACK_DEPTH = 4;

function readerUrl(item: BrowseEntry) {
  return baseUrl(`reader/${encodeURIComponent(item.title_id)}/${encodeURIComponent(item.id)}`);
}

function wrapIndex(index: number, total: number) {
  if (total <= 0) return 0;
  return ((index % total) + total) % total;
}

/**
 * Circular bidirectional stack: shortest wrap distance.
 * previous → left (side -1), next → right (side +1). depth 0 = front.
 * When equidistant, prefer right (next).
 */
function stackPlacement(index: number, active: number, total: number) {
  if (index === active || total <= 1) return { side: 0, depth: 0 };
  const forward = (index - active + total) % total;
  const backward = (active - index + total) % total;
  if (forward <= backward) return { side: 1, depth: forward };
  return { side: -1, depth: backward };
}

export function ContinueCarousel({ items }: { items: BrowseEntry[] }) {
  const { t } = useI18n();
  const [activeIndex, setActiveIndex] = useState(0);
  const multi = items.length > 1;
  const total = items.length;

  const goToIndex = useCallback(
    (index: number) => {
      if (total <= 0) return;
      setActiveIndex(wrapIndex(index, total));
    },
    [total],
  );

  const step = (dir: -1 | 1) => {
    goToIndex(activeIndex + dir);
  };

  const onKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
    if (!multi) return;
    if (event.key === 'ArrowLeft') {
      event.preventDefault();
      step(-1);
    } else if (event.key === 'ArrowRight') {
      event.preventDefault();
      step(1);
    }
  };

  return (
    <section
      className="mango-browse-section"
      aria-roledescription="carousel"
      aria-label={t('continueReading')}
    >
      <h2>{t('continueReading')}</h2>
      <div
        className={`mango-continue-stack${multi ? '' : ' mango-continue-stack--single'}`}
        tabIndex={multi ? 0 : -1}
        onKeyDown={onKeyDown}
      >
        {multi ? (
          <button
            type="button"
            className="mango-continue-stack__arrow mango-continue-stack__arrow--prev mango-btn mango-btn--icon"
            aria-label={t('previousEntry')}
            onClick={() => step(-1)}
          >
            <Icon icon={icons.back} size={18} />
          </button>
        ) : null}

        <div className="mango-continue-stack__stage">
          {items.map((item, index) => {
            const active = index === activeIndex;
            const { side, depth } = stackPlacement(index, activeIndex, total);
            if (!active && depth > MAX_STACK_DEPTH) return null;

            return (
              <article
                key={item.id}
                className={`mango-continue-stack__card${active ? ' mango-continue-stack__card--active' : ' mango-continue-stack__card--back'}${side < 0 ? ' mango-continue-stack__card--left' : ''}${side > 0 ? ' mango-continue-stack__card--right' : ''}`}
                style={
                  {
                    '--stack-side': side,
                    '--stack-depth': depth,
                    zIndex: total - depth,
                  } as CSSProperties
                }
                aria-current={active ? 'true' : undefined}
                aria-hidden={active ? undefined : true}
              >
                {active ? (
                  <div className="mango-continue-stack__face">
                    <div className="mango-continue-stack__cover" aria-hidden="true">
                      {item.cover_url ? (
                        <img src={item.cover_url} alt="" />
                      ) : (
                        <div className="mango-card__placeholder" />
                      )}
                    </div>
                    <div className="mango-continue-stack__meta">
                      <h3>{item.name}</h3>
                      <p>
                        {item.page > 0
                          ? `${item.page} / ${item.pages} ${t('page')}`
                          : `${item.pages} ${t('page')}`}
                      </p>
                      <ProgressBar value={item.progress} />
                      <div className="mango-actions">
                        <a className="mango-btn mango-btn--primary" href={readerUrl(item)}>
                          <Icon icon={icons.continue} size={16} />
                          {t('continue')}
                        </a>
                      </div>
                    </div>
                  </div>
                ) : (
                  <button
                    type="button"
                    className="mango-continue-stack__back"
                    tabIndex={-1}
                    onClick={() => goToIndex(index)}
                    aria-label={item.name}
                  >
                    {item.cover_url ? (
                      <img src={item.cover_url} alt="" />
                    ) : (
                      <div className="mango-card__placeholder" />
                    )}
                  </button>
                )}
              </article>
            );
          })}
        </div>

        {multi ? (
          <button
            type="button"
            className="mango-continue-stack__arrow mango-continue-stack__arrow--next mango-btn mango-btn--icon"
            aria-label={t('nextEntry')}
            onClick={() => step(1)}
          >
            <Icon icon={icons.forward} size={18} />
          </button>
        ) : null}
      </div>
    </section>
  );
}
