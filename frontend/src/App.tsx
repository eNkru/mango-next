import { lazy, Suspense } from 'react';
import { Navigate, Route, Routes, useParams, useSearchParams } from 'react-router-dom';
import { HomePage } from './pages/HomePage';
import { LoginPage } from './pages/LoginPage';
import { LibraryPage } from './pages/LibraryPage';
import { AppShell } from './shell/AppShell';
import { useI18n } from './lib/i18n';
import { ErrorState, LoadingState } from './shell/StatePanels';

// Heavy, route-gated pages are lazy-loaded so the initial main.js (login /
// home / library) does not ship the reader, admin, or title-detail code.
// Each page uses a named export, so the adapter re-exports it as default.
const TitleDetailPage = lazy(() =>
  import('./pages/TitleDetailPage').then((m) => ({ default: m.TitleDetailPage })),
);
const TagDetailPage = lazy(() =>
  import('./pages/TagDetailPage').then((m) => ({ default: m.TagDetailPage })),
);
const TagsIndexPage = lazy(() =>
  import('./pages/TagsIndexPage').then((m) => ({ default: m.TagsIndexPage })),
);
const AdminPage = lazy(() => import('./pages/AdminPage').then((m) => ({ default: m.AdminPage })));
const UserListPage = lazy(() =>
  import('./pages/UserListPage').then((m) => ({ default: m.UserListPage })),
);
const UserEditPage = lazy(() =>
  import('./pages/UserEditPage').then((m) => ({ default: m.UserEditPage })),
);
const MissingItemsPage = lazy(() =>
  import('./pages/MissingItemsPage').then((m) => ({ default: m.MissingItemsPage })),
);
const ReaderPage = lazy(() =>
  import('./pages/reader/ReaderPage').then((m) => ({ default: m.ReaderPage })),
);

function RouteFallback() {
  const { t } = useI18n();
  return <LoadingState message={t('loading')} />;
}

function TitleDetailRoute() {
  const { titleId = '' } = useParams();
  return <TitleDetailPage titleId={decodeURIComponent(titleId)} />;
}

function TagDetailRoute() {
  const { tag = '' } = useParams();
  const [search] = useSearchParams();
  return (
    <TagDetailPage
      tag={decodeURIComponent(tag)}
      showHidden={search.get('show_hidden') === '1'}
    />
  );
}

function ReaderRoute() {
  const { tid = '', eid = '', page } = useParams();
  const pageNum = page ? Number(page) : undefined;
  return (
    <ReaderPage
      tid={decodeURIComponent(tid)}
      eid={decodeURIComponent(eid)}
      initialPage={pageNum && Number.isFinite(pageNum) && pageNum >= 1 ? pageNum : undefined}
    />
  );
}

function UserEditRoute() {
  const [search] = useSearchParams();
  return <UserEditPage username={search.get('username') ?? undefined} />;
}

function UnknownPage() {
  const { t } = useI18n();
  return (
    <AppShell title={t('unknownPage')} subtitle="">
      <ErrorState message={t('unknownPageMessage', { pageId: 'unknown' })} />
    </AppShell>
  );
}

export function App() {
  return (
    <Suspense fallback={<RouteFallback />}>
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/library" element={<LibraryPage />} />
        <Route path="/book/:titleId" element={<TitleDetailRoute />} />
        <Route path="/tags" element={<TagsIndexPage />} />
        <Route path="/tags/:tag" element={<TagDetailRoute />} />
        <Route path="/admin" element={<AdminPage />} />
        <Route path="/admin/user" element={<UserListPage />} />
        <Route path="/admin/user/edit" element={<UserEditRoute />} />
        <Route path="/admin/missing" element={<MissingItemsPage />} />
        <Route path="/reader/:tid/:eid" element={<ReaderRoute />} />
        <Route path="/reader/:tid/:eid/:page" element={<ReaderRoute />} />
        <Route path="*" element={<UnknownPage />} />
        {/* keep Navigate available for future redirects */}
        <Route path="/home" element={<Navigate to="/" replace />} />
      </Routes>
    </Suspense>
  );
}
