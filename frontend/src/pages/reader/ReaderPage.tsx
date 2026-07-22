import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { baseUrl } from '../../lib/baseUrl';
import { AppLink, useAppNavigate } from '../../lib/AppLink';
import { useI18n } from '../../lib/i18n';
import { AlertHost } from '../../shell/AlertHost';
import { Icon } from '../../shell/Icon';
import { icons } from '../../shell/icons';
import { ErrorState, LoadingState } from '../../shell/StatePanels';
import { isLongPageTitle } from './readerMath';
import { ReaderControls } from './ReaderControls';
import { ReaderTopBar } from './ReaderTopBar';
import { ReaderViewport } from './ReaderViewport';
import { useReaderBootstrap } from './useReaderBootstrap';
import { useReaderNavigation } from './useReaderNavigation';
import { useReaderPrefs } from './useReaderPrefs';
import { useReaderProgress } from './useReaderProgress';

const EDGE_PX = 36;
const IDLE_MS = 1800;

export function ReaderPage({
  tid,
  eid,
  initialPage = 1,
}: {
  tid: string;
  eid: string;
  initialPage?: number;
}) {
  const { t } = useI18n();
  const navigate = useAppNavigate();

  const { state } = useReaderBootstrap(tid, eid);
  const { prefs, setPrefs } = useReaderPrefs();

  const [barVisible, setBarVisible] = useState(false);
  const [controlsOpen, setControlsOpen] = useState(false);
  const [flipSide, setFlipSide] = useState<'left' | 'right' | null>(null);
  const hideTimer = useRef<number | null>(null);
  const pointerInBar = useRef(false);

  const data = state.status === 'ready' ? state.data : null;
  const pages = data?.entry.pages ?? 0;
  const longPages = useMemo(
    () => (data ? isLongPageTitle(data.dimensions) : false),
    [data],
  );

  const progress = useReaderProgress({
    tid,
    eid,
    pages,
    longPages,
    initialSaved: data?.entry.progress ?? 0,
  });

  const onPageChange = useCallback(
    (page: number) => {
      void progress.save(page);
    },
    [progress],
  );

  const nav = useReaderNavigation({
    tid,
    eid,
    pages: pages || 1,
    initialPage,
    enableRightToLeft: prefs.enableRightToLeft,
    preloadLookahead: prefs.preloadLookahead,
    mode: prefs.mode,
    onPageChange,
  });

  const clearHideTimer = () => {
    if (hideTimer.current != null) {
      window.clearTimeout(hideTimer.current);
      hideTimer.current = null;
    }
  };

  const scheduleHide = useCallback(() => {
    clearHideTimer();
    hideTimer.current = window.setTimeout(() => {
      if (!pointerInBar.current && !controlsOpen) setBarVisible(false);
    }, IDLE_MS);
  }, [controlsOpen]);

  const showBar = useCallback(
    (intentional = false) => {
      setBarVisible(true);
      if (!intentional) scheduleHide();
      else clearHideTimer();
    },
    [scheduleHide],
  );

  useEffect(() => {
    document.title = data ? `Mango - ${data.entry.name}` : `Mango - ${t('reader')}`;
  }, [data, t]);

  useEffect(() => {
    const onMove = (event: PointerEvent) => {
      if (event.clientY <= EDGE_PX) showBar(false);
    };
    const onKey = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        if (controlsOpen) {
          setControlsOpen(false);
          scheduleHide();
        } else {
          showBar(true);
          setControlsOpen(true);
        }
      }
    };
    window.addEventListener('pointermove', onMove);
    window.addEventListener('keydown', onKey);
    return () => {
      window.removeEventListener('pointermove', onMove);
      window.removeEventListener('keydown', onKey);
      clearHideTimer();
    };
  }, [controlsOpen, scheduleHide, showBar]);

  const openControls = (page?: number) => {
    if (page) nav.setPage(page, { replace: true });
    setControlsOpen(true);
    showBar(true);
  };

  const closeControls = () => {
    setControlsOpen(false);
    scheduleHide();
  };

  const toAppPath = (url: string) => {
    if (!url) return 'library';
    try {
      const u = url.startsWith('http') ? new URL(url) : new URL(url, window.location.origin);
      let path = u.pathname + u.search;
      const base = baseUrl().replace(/\/$/, '');
      if (base && base !== '' && path.startsWith(base)) {
        path = path.slice(base.length) || '/';
      }
      return path.replace(/^\//, '');
    } catch {
      return url.replace(/^\//, '');
    }
  };

  const completeAndGo = async (url: string) => {
    await progress.complete();
    navigate(toAppPath(url), { replace: true });
  };

  const onZoneClick = (zoneIsRight: boolean) => {
    const before = nav.page;
    const ok = nav.flipWithRtl(zoneIsRight);
    if (ok === false) {
      openControls(before);
      return;
    }
    if (prefs.enableFlipAnimation) {
      // Legacy click zones: left zone → left anim, right zone → right anim.
      setFlipSide(zoneIsRight ? 'right' : 'left');
      window.setTimeout(() => setFlipSide(null), 400);
    }
  };

  if (state.status === 'loading') {
    return (
      <div className="mango-reader mango-reader--gate">
        <LoadingState message={t('loading')} />
        <AlertHost />
      </div>
    );
  }

  if (state.status === 'error' || !data) {
    return (
      <div className="mango-reader mango-reader--gate">
        <ErrorState message={state.status === 'error' ? state.message : t('readerError')} />
        <p>
          <AppLink className="mango-btn" to="library">
            <Icon icon={icons.library} size={16} />
            {t('library')}
          </AppLink>
        </p>
        <AlertHost />
      </div>
    );
  }

  const exitPath = data.exit_url ? toAppPath(data.exit_url) : `book/${encodeURIComponent(tid)}`;

  return (
    <div className="mango-reader">
      <div
        className="mango-reader-edge"
        onPointerEnter={() => showBar(false)}
        aria-hidden
      />
      <ReaderTopBar
        visible={barVisible || controlsOpen}
        title={data.title.name}
        entryName={data.entry.name}
        page={nav.page}
        pages={pages}
        exitUrl={exitPath}
        onOpenControls={() => openControls(nav.page)}
        onPointerEnter={() => {
          pointerInBar.current = true;
          showBar(true);
        }}
        onPointerLeave={() => {
          pointerInBar.current = false;
          scheduleHide();
        }}
      />
      <ReaderViewport
        tid={tid}
        eid={eid}
        pages={pages}
        page={nav.page}
        dimensions={data.dimensions}
        mode={prefs.mode}
        margin={prefs.margin}
        fitType={prefs.fitType}
        enableFlipAnimation={prefs.enableFlipAnimation}
        flipSide={flipSide}
        onImageClick={(p) => openControls(p)}
        onZoneClick={onZoneClick}
        onVisiblePage={(p) => nav.setPage(p)}
      />
      {prefs.mode === 'continuous' ? (
        <div className="mango-reader-footer">
          {data.next_entry_url ? (
            <button
              type="button"
              className="mango-btn mango-btn--primary"
              onClick={() => void completeAndGo(data.next_entry_url || exitPath)}
            >
              <Icon icon={icons.play} size={16} />
              {t('nextEntry')}
            </button>
          ) : (
            <button type="button" className="mango-btn mango-btn--primary" onClick={() => void completeAndGo(exitPath)}>
              <Icon icon={icons.exit} size={16} />
              {t('exitReader')}
            </button>
          )}
        </div>
      ) : null}
      <ReaderControls
        open={controlsOpen}
        entryName={data.entry.name}
        entryId={data.entry.id}
        page={nav.page}
        pages={pages}
        entries={data.entries}
        prefs={prefs}
        nextEntryUrl={data.next_entry_url}
        previousEntryUrl={data.previous_entry_url}
        exitUrl={exitPath}
        onClose={closeControls}
        onJumpPage={(p) => {
          nav.setPage(p);
          closeControls();
        }}
        onPrefs={setPrefs}
        onJumpEntry={(nextEid) => {
          navigate(`reader/${encodeURIComponent(tid)}/${encodeURIComponent(nextEid)}`, { replace: true });
        }}
        onPreviousEntry={
          data.previous_entry_url
            ? () => void completeAndGo(data.previous_entry_url)
            : undefined
        }
        onNextEntry={() => void completeAndGo(data.next_entry_url || exitPath)}
        onExit={() => void completeAndGo(exitPath)}
      />
      <AlertHost />
    </div>
  );
}
