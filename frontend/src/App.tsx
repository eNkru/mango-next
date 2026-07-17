import { readBoot } from './lib/boot';
import { PlaceholderPage } from './pages/PlaceholderPage';
import { AppShell } from './shell/AppShell';
import { ErrorState } from './shell/StatePanels';

export function App() {
  const boot = readBoot();

  switch (boot.pageId) {
    case 'react-preview':
      return <PlaceholderPage />;
    default:
      return (
        <AppShell title="未知页面" subtitle={boot.pageId}>
          <ErrorState message={`No React page registered for pageId=${boot.pageId}`} />
        </AppShell>
      );
  }
}
