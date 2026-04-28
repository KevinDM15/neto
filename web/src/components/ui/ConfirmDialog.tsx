interface ConfirmDialogProps {
  summary: string;
  pendingTool: Record<string, unknown>;
  onConfirm: (pendingTool: Record<string, unknown>) => void;
  onCancel: () => void;
}

export function ConfirmDialog({ summary, pendingTool, onConfirm, onCancel }: ConfirmDialogProps) {
  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        backgroundColor: 'rgba(0,0,0,0.5)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 50,
      }}
    >
      <div
        style={{
          backgroundColor: '#fff',
          borderRadius: '10px',
          padding: '1.5rem',
          maxWidth: '420px',
          width: '90%',
          display: 'flex',
          flexDirection: 'column',
          gap: '1rem',
        }}
      >
        <h2 style={{ margin: 0, fontSize: '1rem', fontWeight: 700 }}>Confirmar acción</h2>
        <p style={{ margin: 0, fontSize: '0.875rem', color: '#374151' }}>{summary}</p>
        <div style={{ display: 'flex', gap: '0.75rem', justifyContent: 'flex-end' }}>
          <button
            onClick={onCancel}
            style={{
              padding: '0.5rem 1rem',
              border: '1px solid #d1d5db',
              borderRadius: '6px',
              background: '#fff',
              cursor: 'pointer',
              fontSize: '0.875rem',
            }}
          >
            Cancelar
          </button>
          <button
            onClick={() => onConfirm(pendingTool)}
            style={{
              padding: '0.5rem 1rem',
              backgroundColor: '#dc2626',
              color: '#fff',
              border: 'none',
              borderRadius: '6px',
              cursor: 'pointer',
              fontSize: '0.875rem',
              fontWeight: 600,
            }}
          >
            Confirmar
          </button>
        </div>
      </div>
    </div>
  );
}
