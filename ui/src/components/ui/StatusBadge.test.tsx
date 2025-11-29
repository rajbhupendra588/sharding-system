import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import StatusBadge from './StatusBadge';

describe('StatusBadge', () => {
  it('should render status badge with status text', () => {
    render(<StatusBadge status="healthy" />);
    expect(screen.getByText('Healthy')).toBeInTheDocument();
  });

  it('should capitalize status text', () => {
    render(<StatusBadge status="active" />);
    expect(screen.getByText('Active')).toBeInTheDocument();
  });

  it('should handle lowercase status', () => {
    render(<StatusBadge status="degraded" />);
    expect(screen.getByText('Degraded')).toBeInTheDocument();
  });

  it('should handle uppercase status', () => {
    render(<StatusBadge status="FAILED" />);
    // StatusBadge capitalizes first letter only: "FAILED" -> "FAILED" (F uppercase, rest stays)
    expect(screen.getByText('FAILED')).toBeInTheDocument();
  });

  it('should handle mixed case status', () => {
    render(<StatusBadge status="MiGrAtInG" />);
    // StatusBadge only capitalizes first letter, so "MiGrAtInG" becomes "MiGrAtInG"
    expect(screen.getByText('MiGrAtInG')).toBeInTheDocument();
  });

  it('should apply custom className', () => {
    render(<StatusBadge status="healthy" className="custom-badge-class" />);
    const badge = screen.getByText('Healthy');
    expect(badge.className).toContain('custom-badge-class');
  });

  it('should have correct base classes', () => {
    render(<StatusBadge status="pending" />);
    const badge = screen.getByText('Pending');
    expect(badge.className).toContain('inline-flex');
    expect(badge.className).toContain('items-center');
    expect(badge.className).toContain('px-2.5');
    expect(badge.className).toContain('py-0.5');
    expect(badge.className).toContain('rounded-full');
    expect(badge.className).toContain('text-xs');
    expect(badge.className).toContain('font-medium');
    expect(badge.className).toContain('border');
  });

  it('should render different statuses correctly', () => {
    const statuses = ['healthy', 'degraded', 'unhealthy', 'pending', 'migrating'];
    
    statuses.forEach((status) => {
      const { unmount } = render(<StatusBadge status={status} />);
      expect(screen.getByText(status.charAt(0).toUpperCase() + status.slice(1))).toBeInTheDocument();
      unmount();
    });
  });
});

