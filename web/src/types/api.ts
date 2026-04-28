export interface Account {
  id: string;
  name: string;
  type: string;
  balance: number;
  currency: string;
  created_at: string;
}

export interface Category {
  id: string;
  name: string;
  type: 'income' | 'expense';
}

export interface Transaction {
  id: string;
  account_id: string;
  category_id: string | null;
  amount: number;
  currency: string;
  description: string;
  date: string;
  created_at: string;
}

export interface PendingConfirmation {
  tool_name: string;
  summary: string;
  pending_tool: Record<string, unknown>;
}

export interface ChatRequest {
  conversation_id: string | null;
  message: string;
  confirm: boolean;
  pending_tool: Record<string, unknown> | null;
}

export interface ChatResponse {
  conversation_id: string;
  reply: string;
  pending_confirmation?: PendingConfirmation;
}
