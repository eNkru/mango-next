import { Navigate, Route, Routes, useParams, useSearchParams } from 'react-router-dom';
import { LoginPage } from './pages/LoginPage';
import { HomePage } from './pages/HomePage';
import { LibraryPage } from './pages/LibraryPage';
import { MissingItemsPage } from './pages/MissingItemsPage';
import { TagDetailPage } from './pages/TagDetailPage';
import { TagsIndexPage } from './pages/TagsIndexPage';
import { UserEditPage } from './pages/UserEditPage';
import { UserListPage } from './pages/UserListPage';
import { TitleDetailPage } from './pages/TitleDetailPage';
import { AdminPage } from './pages/AdminPage';
import { ReaderPage } from './pages/reader/ReaderPage';
import { AppShell } from './shell/AppShell';
import { useI18n } from './lib/i18n';
import { ErrorState } from './shell/StatePanels';

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
  );
}
