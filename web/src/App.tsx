import { useSession } from './hooks/useSession';
import { AppShell } from './components/layout/AppShell';
import { LoginPage } from './components/chat/LoginPage';
import { ChatPage } from './components/chat/ChatPage';

export default function App() {
  const { session, loading } = useSession();

  if (loading) {
    return (
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100vh' }}>
        <span style={{ color: '#6b7280', fontSize: '0.875rem' }}>Cargando…</span>
      </div>
    );
  }

  if (!session) {
    return <LoginPage />;
  }

  return (
    <AppShell>
      <ChatPage />
    </AppShell>
  );
}
