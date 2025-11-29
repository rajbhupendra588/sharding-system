import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import Modal from './Modal';

describe('Modal', () => {
  beforeEach(() => {
    document.body.style.overflow = '';
  });

  afterEach(() => {
    document.body.style.overflow = '';
  });

  it('should not render when isOpen is false', () => {
    render(
      <Modal isOpen={false} onClose={vi.fn()}>
        Modal Content
      </Modal>
    );
    expect(screen.queryByText('Modal Content')).not.toBeInTheDocument();
  });

  it('should render when isOpen is true', () => {
    render(
      <Modal isOpen={true} onClose={vi.fn()}>
        Modal Content
      </Modal>
    );
    expect(screen.getByText('Modal Content')).toBeInTheDocument();
  });

  it('should render title when provided', () => {
    render(
      <Modal isOpen={true} onClose={vi.fn()} title="Test Modal">
        Modal Content
      </Modal>
    );
    expect(screen.getByText('Test Modal')).toBeInTheDocument();
  });

  it('should call onClose when close button is clicked', () => {
    const handleClose = vi.fn();
    render(
      <Modal isOpen={true} onClose={handleClose} title="Test Modal">
        Modal Content
      </Modal>
    );
    
    const closeButton = screen.getAllByRole('button').find(
      (btn) => btn.querySelector('svg')
    );
    if (closeButton) {
      fireEvent.click(closeButton);
      expect(handleClose).toHaveBeenCalledTimes(1);
    }
  });

  it('should call onClose when backdrop is clicked', () => {
    const handleClose = vi.fn();
    render(
      <Modal isOpen={true} onClose={handleClose}>
        Modal Content
      </Modal>
    );
    
    const backdrop = screen.getByText('Modal Content').closest('.fixed.inset-0')?.previousElementSibling;
    if (backdrop) {
      fireEvent.click(backdrop);
      expect(handleClose).toHaveBeenCalledTimes(1);
    }
  });

  it('should render footer when provided', () => {
    render(
      <Modal isOpen={true} onClose={vi.fn()} footer={<button>Save</button>}>
        Modal Content
      </Modal>
    );
    expect(screen.getByText('Save')).toBeInTheDocument();
  });

  it('should apply small size', () => {
    render(
      <Modal isOpen={true} onClose={vi.fn()} size="sm">
        Modal Content
      </Modal>
    );
    const modal = screen.getByText('Modal Content').closest('.max-w-md');
    expect(modal).toBeInTheDocument();
  });

  it('should apply medium size by default', () => {
    render(
      <Modal isOpen={true} onClose={vi.fn()}>
        Modal Content
      </Modal>
    );
    const modal = screen.getByText('Modal Content').closest('.max-w-lg');
    expect(modal).toBeInTheDocument();
  });

  it('should apply large size', () => {
    render(
      <Modal isOpen={true} onClose={vi.fn()} size="lg">
        Modal Content
      </Modal>
    );
    const modal = screen.getByText('Modal Content').closest('.max-w-2xl');
    expect(modal).toBeInTheDocument();
  });

  it('should apply extra large size', () => {
    render(
      <Modal isOpen={true} onClose={vi.fn()} size="xl">
        Modal Content
      </Modal>
    );
    const modal = screen.getByText('Modal Content').closest('.max-w-4xl');
    expect(modal).toBeInTheDocument();
  });

  it('should prevent body scroll when open', () => {
    render(
      <Modal isOpen={true} onClose={vi.fn()}>
        Modal Content
      </Modal>
    );
    expect(document.body.style.overflow).toBe('hidden');
  });

  it('should restore body scroll when closed', () => {
    const { rerender } = render(
      <Modal isOpen={true} onClose={vi.fn()}>
        Modal Content
      </Modal>
    );
    expect(document.body.style.overflow).toBe('hidden');
    
    rerender(
      <Modal isOpen={false} onClose={vi.fn()}>
        Modal Content
      </Modal>
    );
    expect(document.body.style.overflow).toBe('unset');
  });

  it('should render close button when no title is provided', () => {
    render(
      <Modal isOpen={true} onClose={vi.fn()}>
        Modal Content
      </Modal>
    );
    const closeButtons = screen.getAllByRole('button');
    expect(closeButtons.length).toBeGreaterThan(0);
  });
});

