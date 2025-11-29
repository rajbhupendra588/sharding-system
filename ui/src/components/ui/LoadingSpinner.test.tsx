import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import LoadingSpinner from './LoadingSpinner';

describe('LoadingSpinner', () => {
  it('should render spinner', () => {
    const { container } = render(<LoadingSpinner />);
    const spinner = container.querySelector('svg');
    expect(spinner).toBeInTheDocument();
  });

  it('should apply small size', () => {
    const { container } = render(<LoadingSpinner size="sm" />);
    const spinner = container.querySelector('svg');
    expect(spinner?.getAttribute('class')).toContain('h-4 w-4');
  });

  it('should apply medium size by default', () => {
    const { container } = render(<LoadingSpinner />);
    const spinner = container.querySelector('svg');
    expect(spinner?.getAttribute('class')).toContain('h-8 w-8');
  });

  it('should apply large size', () => {
    const { container } = render(<LoadingSpinner size="lg" />);
    const spinner = container.querySelector('svg');
    expect(spinner?.getAttribute('class')).toContain('h-12 w-12');
  });

  it('should apply custom className', () => {
    const { container } = render(<LoadingSpinner className="custom-spinner-class" />);
    const wrapper = container.firstChild as HTMLElement;
    expect(wrapper?.className).toContain('custom-spinner-class');
  });

  it('should have spinning animation', () => {
    const { container } = render(<LoadingSpinner />);
    const spinner = container.querySelector('svg');
    expect(spinner?.getAttribute('class')).toContain('animate-spin');
  });

  it('should be centered in container', () => {
    const { container } = render(<LoadingSpinner />);
    const wrapper = container.firstChild as HTMLElement;
    expect(wrapper?.className).toContain('flex');
    expect(wrapper?.className).toContain('items-center');
    expect(wrapper?.className).toContain('justify-center');
  });
});

