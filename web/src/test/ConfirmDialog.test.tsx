import { render, screen } from '@testing-library/react';
import { describe, it, vi } from 'vitest';
import { ConfirmDialog } from '../components/ui/ConfirmDialog';

describe('ConfirmDialog', () => {
  it('renders summary text and Confirmar/Cancelar buttons', () => {
    render(
      <ConfirmDialog
        summary="¿Estás seguro de eliminar esta transacción?"
        pendingTool={{ id: '123' }}
        onConfirm={vi.fn()}
        onCancel={vi.fn()}
      />
    );
    expect(screen.getByText('¿Estás seguro de eliminar esta transacción?')).toBeInTheDocument();
    expect(screen.getByText('Confirmar')).toBeInTheDocument();
    expect(screen.getByText('Cancelar')).toBeInTheDocument();
  });
});
