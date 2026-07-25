import { useEffect, type RefObject } from 'react';

const FOCUSABLE =
  'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';

function listFocusable(root: HTMLElement): HTMLElement[] {
  return Array.from(root.querySelectorAll<HTMLElement>(FOCUSABLE)).filter(
    (el) => !el.hasAttribute('disabled') && el.tabIndex !== -1 && el.offsetParent !== null,
  );
}

/**
 * When `active`, trap Tab focus inside `containerRef` and restore focus to
 * `restoreRef` (or the previously focused element) on deactivate.
 */
export function useFocusTrap(
  active: boolean,
  containerRef: RefObject<HTMLElement | null>,
  restoreRef?: RefObject<HTMLElement | null>,
) {
  useEffect(() => {
    if (!active) return;
    const root = containerRef.current;
    if (!root) return;

    const previouslyFocused =
      (document.activeElement instanceof HTMLElement ? document.activeElement : null);

    const focusFirst = () => {
      const items = listFocusable(root);
      (items[0] ?? root).focus();
    };

    // Defer so dialog content is mounted.
    const raf = window.requestAnimationFrame(focusFirst);

    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key !== 'Tab') return;
      const items = listFocusable(root);
      if (!items.length) {
        event.preventDefault();
        root.focus();
        return;
      }
      const first = items[0];
      const last = items[items.length - 1];
      const current = document.activeElement as HTMLElement | null;
      if (event.shiftKey) {
        if (current === first || !root.contains(current)) {
          event.preventDefault();
          last.focus();
        }
      } else if (current === last) {
        event.preventDefault();
        first.focus();
      }
    };

    root.addEventListener('keydown', onKeyDown);
    return () => {
      window.cancelAnimationFrame(raf);
      root.removeEventListener('keydown', onKeyDown);
      const restore = restoreRef?.current ?? previouslyFocused;
      if (restore && document.contains(restore)) {
        restore.focus();
      }
    };
  }, [active, containerRef, restoreRef]);
}
