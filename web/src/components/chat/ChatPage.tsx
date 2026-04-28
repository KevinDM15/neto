import React, { useEffect, useRef, useState } from 'react';
import ReactMarkdown from 'react-markdown';
import { sendChatMessage } from '../../services/api';
import type { PendingConfirmation } from '../../types/api';
import { ConfirmDialog } from '../ui/ConfirmDialog';

interface Message {
  role: 'user' | 'ai';
  content: string;
}

export function ChatPage() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [conversationId, setConversationId] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [pending, setPending] = useState<PendingConfirmation | null>(null);
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (bottomRef.current && typeof bottomRef.current.scrollIntoView === 'function') {
      bottomRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [messages]);

  async function submit(message: string, confirm = false, pendingTool: Record<string, unknown> | null = null) {
    setLoading(true);
    try {
      const res = await sendChatMessage({
        conversation_id: conversationId,
        message,
        confirm,
        pending_tool: pendingTool,
      });
      setConversationId(res.conversation_id);
      setMessages((prev) => [...prev, { role: 'ai', content: res.reply }]);
      if (res.pending_confirmation) {
        setPending(res.pending_confirmation);
      }
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Error desconocido';
      setMessages((prev) => [...prev, { role: 'ai', content: `⚠️ ${msg}` }]);
    } finally {
      setLoading(false);
    }
  }

  async function handleSend() {
    const text = input.trim();
    if (!text || loading) return;
    setMessages((prev) => [...prev, { role: 'user', content: text }]);
    setInput('');
    await submit(text);
  }

  async function handleConfirm(pendingTool: Record<string, unknown>) {
    setPending(null);
    await submit('', true, pendingTool);
  }

  function handleCancel() {
    setPending(null);
  }

  function handleKeyDown(e: React.KeyboardEvent<HTMLTextAreaElement>) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      void handleSend();
    }
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      {/* Message list */}
      <div style={{ flex: 1, overflowY: 'auto', padding: '1rem', display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
        {messages.length === 0 && (
          <p style={{ color: '#9ca3af', textAlign: 'center', marginTop: '3rem', fontSize: '0.875rem' }}>
            Preguntame sobre tus finanzas…
          </p>
        )}
        {messages.map((msg, i) => (
          <div
            key={i}
            style={{
              alignSelf: msg.role === 'user' ? 'flex-end' : 'flex-start',
              maxWidth: '75%',
              padding: '0.625rem 0.875rem',
              borderRadius: '10px',
              backgroundColor: msg.role === 'user' ? '#111827' : '#f3f4f6',
              color: msg.role === 'user' ? '#fff' : '#111827',
              fontSize: '0.875rem',
              lineHeight: 1.6,
            }}
          >
            {msg.role === 'ai' ? (
              <ReactMarkdown>{msg.content}</ReactMarkdown>
            ) : (
              msg.content
            )}
          </div>
        ))}
        {loading && (
          <div
            style={{
              alignSelf: 'flex-start',
              padding: '0.625rem 0.875rem',
              borderRadius: '10px',
              backgroundColor: '#f3f4f6',
              color: '#6b7280',
              fontSize: '0.875rem',
            }}
          >
            Pensando…
          </div>
        )}
        <div ref={bottomRef} />
      </div>

      {/* Input bar */}
      <div
        style={{
          borderTop: '1px solid #e5e7eb',
          padding: '0.75rem 1rem',
          display: 'flex',
          gap: '0.5rem',
          backgroundColor: '#fff',
        }}
      >
        <textarea
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          disabled={loading}
          rows={1}
          placeholder="Escribe un mensaje…"
          style={{
            flex: 1,
            resize: 'none',
            border: '1px solid #d1d5db',
            borderRadius: '6px',
            padding: '0.5rem 0.75rem',
            fontSize: '0.875rem',
            outline: 'none',
            fontFamily: 'inherit',
          }}
        />
        <button
          onClick={() => void handleSend()}
          disabled={loading || !input.trim()}
          style={{
            padding: '0.5rem 1rem',
            backgroundColor: '#111827',
            color: '#fff',
            border: 'none',
            borderRadius: '6px',
            cursor: loading || !input.trim() ? 'not-allowed' : 'pointer',
            fontSize: '0.875rem',
            fontWeight: 600,
            flexShrink: 0,
          }}
        >
          Enviar
        </button>
      </div>

      {pending && (
        <ConfirmDialog
          summary={pending.summary}
          pendingTool={pending.pending_tool}
          onConfirm={handleConfirm}
          onCancel={handleCancel}
        />
      )}
    </div>
  );
}
