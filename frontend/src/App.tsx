import { readBoot } from './lib/boot';
import { MissingItemsPage } from './pages/MissingItemsPage';
import { PlaceholderPage } from './pages/PlaceholderPage';
import { UserEditPage } from './pages/UserEditPage';
import { UserListPage } from './pages/UserListPage';
import { AppShell } from './shell/AppShell';
import { ErrorState } from './shell/StatePanels';

export function App() {
  const boot = readBoot();

  switch (boot.pageId) {
    case 'react-preview':
      return <PlaceholderPage />;
    case 'missing-items':
      return <MissingItemsPage />;
    case 'user-list':
      return <UserListPage />;
    case 'user-edit':
      return <UserEditPage />;
    default:
      return (
        <AppShell title="未知页面" subtitle={boot.pageId}>
          <ErrorState message={`No React page registered for pageId=${boot.pageId}`} />
        </AppShell>
      );
  }
}
