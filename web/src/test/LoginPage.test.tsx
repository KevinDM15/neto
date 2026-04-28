import { render, screen } from '@testing-library/react';
import { describe, it, vi } from 'vitest';
import { LoginPage } from '../components/chat/LoginPage';

vi.mock('../lib/supabase', () => ({
  supabase: {
    auth: {
      signInWithPassword: vi.fn(),
    },
  },
}));

describe('LoginPage', () => {
  it('renders email and password inputs', () => {
    render(<LoginPage />);
    expect(screen.getByPlaceholderText('usuario@ejemplo.com')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('••••••••')).toBeInTheDocument();
  });
});
