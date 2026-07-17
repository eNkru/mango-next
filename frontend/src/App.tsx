import { readBoot } from './lib/boot';
import { MissingItemsPage } from './pages/MissingItemsPage';
import { PlaceholderPage } from './pages/PlaceholderPage';
import { AppShell } from './shell/AppShell';
import { ErrorState } from './shell/StatePanels';

export function App() {
  const boot = readBoot();

  switch (boot.pageId) {
    case 'react-preview':
      return <PlaceholderPage />;
    case 'missing-items':
      return <MissingItemsPage />;
    default:
      return (
        <AppShell title="未知页面" subtitle={boot.pageId}>
          <ErrorState message={`No React page registered for pageId=${boot.pageId}`} />
        </AppShell>
      );
  }
}
