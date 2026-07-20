import { readBoot } from './lib/boot';
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

export function App() {
  const boot = readBoot();
  const { t } = useI18n();

  switch (boot.pageId) {
    case 'login':
      return <LoginPage />;
    case 'home':
      return <HomePage />;
    case 'library':
      return <LibraryPage />;
    case 'title-detail':
      return <TitleDetailPage />;
    case 'reader':
      return <ReaderPage />;
    case 'admin':
      return <AdminPage />;
    case 'missing-items':
      return <MissingItemsPage />;
    case 'user-list':
      return <UserListPage />;
    case 'user-edit':
      return <UserEditPage />;
    case 'tags-index':
      return <TagsIndexPage />;
    case 'tag-detail':
      return <TagDetailPage />;
    default:
      return (
        <AppShell title={t('unknownPage')} subtitle={boot.pageId}>
          <ErrorState message={t('unknownPageMessage', { pageId: boot.pageId })} />
        </AppShell>
      );
  }
}
