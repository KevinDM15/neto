import { render, screen } from '@testing-library/react';
import { describe, it, vi } from 'vitest';
import { ChatPage } from '../components/chat/ChatPage';

vi.mock('../services/api', () => ({
  sendChatMessage: vi.fn(),
}));

vi.mock('../lib/supabase', () => ({
  supabase: {
    auth: {
      getSession: vi.fn().mockResolvedValue({ data: { session: null } }),
    },
  },
}));

describe('ChatPage', () => {
  it('renders the message input textarea', () => {
    render(<ChatPage />);
    expect(screen.getByPlaceholderText('Escribe un mensaje…')).toBeInTheDocument();
  });
});
