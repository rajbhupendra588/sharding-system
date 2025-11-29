import { describe, it, expect } from 'vitest';
import { getStatusColor, getStatusBadgeColor } from './status';

describe('status utilities', () => {
  describe('getStatusColor', () => {
    it('should return green color for healthy status', () => {
      const color = getStatusColor('healthy');
      expect(color).toContain('green');
    });

    it('should return green color for active status', () => {
      const color = getStatusColor('active');
      expect(color).toContain('green');
    });

    it('should return yellow color for degraded status', () => {
      const color = getStatusColor('degraded');
      expect(color).toContain('yellow');
    });

    it('should return red color for unhealthy status', () => {
      const color = getStatusColor('unhealthy');
      expect(color).toContain('red');
    });

    it('should return red color for failed status', () => {
      const color = getStatusColor('failed');
      expect(color).toContain('red');
    });

    it('should return gray color for pending status', () => {
      const color = getStatusColor('pending');
      expect(color).toContain('gray');
    });

    it('should return blue color for migrating status', () => {
      const color = getStatusColor('migrating');
      expect(color).toContain('blue');
    });

    it('should return orange color for readonly status', () => {
      const color = getStatusColor('readonly');
      expect(color).toContain('orange');
    });

    it('should return gray color for inactive status', () => {
      const color = getStatusColor('inactive');
      expect(color).toContain('gray');
    });

    it('should return green color for completed status', () => {
      const color = getStatusColor('completed');
      expect(color).toContain('green');
    });

    it('should return blue color for precopy status', () => {
      const color = getStatusColor('precopy');
      expect(color).toContain('blue');
    });

    it('should return blue color for deltasync status', () => {
      const color = getStatusColor('deltasync');
      expect(color).toContain('blue');
    });

    it('should return purple color for cutover status', () => {
      const color = getStatusColor('cutover');
      expect(color).toContain('purple');
    });

    it('should return indigo color for validation status', () => {
      const color = getStatusColor('validation');
      expect(color).toContain('indigo');
    });

    it('should be case-insensitive', () => {
      expect(getStatusColor('HEALTHY')).toBe(getStatusColor('healthy'));
      expect(getStatusColor('Active')).toBe(getStatusColor('active'));
      expect(getStatusColor('FAILED')).toBe(getStatusColor('failed'));
    });

    it('should return default gray color for unknown status', () => {
      const color = getStatusColor('unknown-status');
      expect(color).toContain('gray');
    });

    it('should include dark mode classes', () => {
      const color = getStatusColor('healthy');
      expect(color).toContain('dark:');
    });
  });

  describe('getStatusBadgeColor', () => {
    it('should return green badge color for healthy status', () => {
      const color = getStatusBadgeColor('healthy');
      expect(color).toContain('green');
    });

    it('should return green badge color for active status', () => {
      const color = getStatusBadgeColor('active');
      expect(color).toContain('green');
    });

    it('should return yellow badge color for degraded status', () => {
      const color = getStatusBadgeColor('degraded');
      expect(color).toContain('yellow');
    });

    it('should return red badge color for unhealthy status', () => {
      const color = getStatusBadgeColor('unhealthy');
      expect(color).toContain('red');
    });

    it('should return red badge color for failed status', () => {
      const color = getStatusBadgeColor('failed');
      expect(color).toContain('red');
    });

    it('should return gray badge color for pending status', () => {
      const color = getStatusBadgeColor('pending');
      expect(color).toContain('gray');
    });

    it('should return blue badge color for migrating status', () => {
      const color = getStatusBadgeColor('migrating');
      expect(color).toContain('blue');
    });

    it('should return orange badge color for readonly status', () => {
      const color = getStatusBadgeColor('readonly');
      expect(color).toContain('orange');
    });

    it('should return gray badge color for inactive status', () => {
      const color = getStatusBadgeColor('inactive');
      expect(color).toContain('gray');
    });

    it('should return green badge color for completed status', () => {
      const color = getStatusBadgeColor('completed');
      expect(color).toContain('green');
    });

    it('should return blue badge color for precopy status', () => {
      const color = getStatusBadgeColor('precopy');
      expect(color).toContain('blue');
    });

    it('should return blue badge color for deltasync status', () => {
      const color = getStatusBadgeColor('deltasync');
      expect(color).toContain('blue');
    });

    it('should return purple badge color for cutover status', () => {
      const color = getStatusBadgeColor('cutover');
      expect(color).toContain('purple');
    });

    it('should return indigo badge color for validation status', () => {
      const color = getStatusBadgeColor('validation');
      expect(color).toContain('indigo');
    });

    it('should be case-insensitive', () => {
      expect(getStatusBadgeColor('HEALTHY')).toBe(getStatusBadgeColor('healthy'));
      expect(getStatusBadgeColor('Active')).toBe(getStatusBadgeColor('active'));
      expect(getStatusBadgeColor('FAILED')).toBe(getStatusBadgeColor('failed'));
    });

    it('should return default gray badge color for unknown status', () => {
      const color = getStatusBadgeColor('unknown-status');
      expect(color).toContain('gray');
    });

    it('should include border classes', () => {
      const color = getStatusBadgeColor('healthy');
      expect(color).toContain('border');
    });

    it('should include dark mode classes', () => {
      const color = getStatusBadgeColor('healthy');
      expect(color).toContain('dark:');
    });

    it('should return different format than getStatusColor', () => {
      const badgeColor = getStatusBadgeColor('healthy');
      const statusColor = getStatusColor('healthy');
      
      // Badge color should include bg- prefix and border
      expect(badgeColor).toContain('bg-');
      expect(badgeColor).toContain('border');
      
      // Status color should include text- and bg- but different structure
      expect(statusColor).toContain('text-');
    });
  });
});

