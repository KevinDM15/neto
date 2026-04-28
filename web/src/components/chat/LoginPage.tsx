import React, { useState } from 'react';
import { supabase } from '../../lib/supabase';

export function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    const { error: authError } = await supabase.auth.signInWithPassword({ email, password });
    if (authError) setError(authError.message);
    setLoading(false);
  }

  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        height: '100vh',
        backgroundColor: '#f9fafb',
      }}
    >
      <form
        onSubmit={handleSubmit}
        style={{
          backgroundColor: '#fff',
          padding: '2rem',
          borderRadius: '10px',
          border: '1px solid #e5e7eb',
          width: '100%',
          maxWidth: '360px',
          display: 'flex',
          flexDirection: 'column',
          gap: '1rem',
        }}
      >
        <h1 style={{ margin: 0, fontSize: '1.5rem', fontWeight: 700 }}>Neto</h1>
        <p style={{ margin: 0, color: '#6b7280', fontSize: '0.875rem' }}>
          Ingresa con tu cuenta para continuar.
        </p>

        <label style={{ display: 'flex', flexDirection: 'column', gap: '4px', fontSize: '0.875rem' }}>
          Correo electrónico
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            style={inputStyle}
            placeholder="usuario@ejemplo.com"
          />
        </label>

        <label style={{ display: 'flex', flexDirection: 'column', gap: '4px', fontSize: '0.875rem' }}>
          Contraseña
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            style={inputStyle}
            placeholder="••••••••"
          />
        </label>

        {error && (
          <p style={{ color: '#dc2626', fontSize: '0.875rem', margin: 0 }}>{error}</p>
        )}

        <button
          type="submit"
          disabled={loading}
          style={{
            padding: '0.625rem',
            backgroundColor: '#111827',
            color: '#fff',
            border: 'none',
            borderRadius: '6px',
            cursor: loading ? 'not-allowed' : 'pointer',
            fontSize: '0.875rem',
            fontWeight: 600,
          }}
        >
          {loading ? 'Ingresando…' : 'Ingresar'}
        </button>
      </form>
    </div>
  );
}

const inputStyle: React.CSSProperties = {
  padding: '0.5rem 0.75rem',
  border: '1px solid #d1d5db',
  borderRadius: '6px',
  fontSize: '0.875rem',
  outline: 'none',
};
