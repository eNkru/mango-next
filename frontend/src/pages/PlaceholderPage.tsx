import { useMemo, useState } from 'react';
import { readBoot } from '../lib/boot';
import { baseUrl } from '../lib/baseUrl';
import { AppShell } from '../shell/AppShell';
import { ConfirmDialog } from '../shell/ConfirmDialog';
import { EmptyState, ErrorState, LoadingState } from '../shell/StatePanels';
import { pushAlert } from '../shell/AlertHost';

type DemoState = 'ready' | 'loading' | 'empty' | 'error';

export function PlaceholderPage() {
  const boot = useMemo(() => readBoot(), []);
  const [demo, setDemo] = useState<DemoState>('ready');
  const [confirmOpen, setConfirmOpen] = useState(false);

  return (
    <AppShell title="React 基础壳" subtitle="验证 Vite 产物、BaseURL 注入与双主题 shell 原语">
      <section className="mango-panel">
        <dl className="mango-meta">
          <div>
            <dt>BaseURL</dt>
            <dd>
              <code>{boot.baseUrl}</code>
            </dd>
          </div>
          <div>
            <dt>Page ID</dt>
            <dd>
              <code>{boot.pageId}</code>
            </dd>
          </div>
          <div>
            <dt>Library link</dt>
            <dd>
              <a href={baseUrl('library')}>{baseUrl('library')}</a>
            </dd>
          </div>
          <div>
            <dt>Theme classes</dt>
            <dd>
              <code>{document.documentElement.className || '(none)'}</code>
            </dd>
          </div>
        </dl>

        <div className="mango-actions">
          <button type="button" className="mango-btn mango-btn--primary" onClick={() => setDemo('ready')}>
            Ready
          </button>
          <button type="button" className="mango-btn" onClick={() => setDemo('loading')}>
            Loading
          </button>
          <button type="button" className="mango-btn" onClick={() => setDemo('empty')}>
            Empty
          </button>
          <button type="button" className="mango-btn" onClick={() => setDemo('error')}>
            Error
          </button>
          <button
            type="button"
            className="mango-btn"
            onClick={() => pushAlert('Toast from React shell', 'success')}
          >
            Toast
          </button>
          <button type="button" className="mango-btn mango-btn--danger" onClick={() => setConfirmOpen(true)}>
            Confirm
          </button>
        </div>
      </section>

      <section className="mango-panel" style={{ marginTop: '1rem' }}>
        {demo === 'ready' ? (
          <p>
            React foundation is mounted. Unmigrated Go templates remain available for other routes.
            Comic/flat and light/dark markers come from the Go HTML shell FOUC script.
          </p>
        ) : null}
        {demo === 'loading' ? <LoadingState /> : null}
        {demo === 'empty' ? <EmptyState message="示例空状态" /> : null}
        {demo === 'error' ? <ErrorState message="示例错误状态" /> : null}
      </section>

      <ConfirmDialog
        open={confirmOpen}
        title="确认示例"
        message="这是 React shell 的确认对话框原语。"
        onCancel={() => setConfirmOpen(false)}
        onConfirm={() => {
          setConfirmOpen(false);
          pushAlert('Confirmed', 'info');
        }}
      />
    </AppShell>
  );
}
