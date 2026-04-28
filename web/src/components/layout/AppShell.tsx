import type { ReactNode } from 'react';
import { supabase } from '../../lib/supabase';

interface AppShellProps {
  children: ReactNode;
}

export function AppShell({ children }: AppShellProps) {
  async function handleSignOut() {
    await supabase.auth.signOut();
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100vh' }}>
      <header
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '0 1.5rem',
          height: '56px',
          borderBottom: '1px solid #e5e7eb',
          backgroundColor: '#fff',
          flexShrink: 0,
        }}
      >
        <span style={{ fontWeight: 700, fontSize: '1.25rem', color: '#111827' }}>Neto</span>
        <button
          onClick={handleSignOut}
          style={{
            padding: '0.375rem 0.875rem',
            border: '1px solid #d1d5db',
            borderRadius: '6px',
            background: '#fff',
            cursor: 'pointer',
            fontSize: '0.875rem',
            color: '#374151',
          }}
        >
          Cerrar sesión
        </button>
      </header>
      <main style={{ flex: 1, overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
        {children}
      </main>
    </div>
  );
}
