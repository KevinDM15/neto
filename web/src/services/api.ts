import { supabase } from '../lib/supabase';
import type { ChatRequest, ChatResponse } from '../types/api';

const API_URL = import.meta.env.VITE_API_URL as string;

async function getAuthHeader(): Promise<string> {
  const { data } = await supabase.auth.getSession();
  const token = data.session?.access_token;
  if (!token) throw new Error('No active session');
  return `Bearer ${token}`;
}

export async function sendChatMessage(req: ChatRequest): Promise<ChatResponse> {
  const auth = await getAuthHeader();
  const response = await fetch(`${API_URL}/api/v1/chat`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: auth,
    },
    body: JSON.stringify(req),
  });

  if (!response.ok) {
    throw new Error(`API error: ${response.status}`);
  }

  return response.json() as Promise<ChatResponse>;
}
