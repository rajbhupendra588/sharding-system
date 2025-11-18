import { describe, it, expect } from 'vitest';
import { formatBytes, formatDuration, formatDate, formatRelativeTime } from './formatting';

describe('formatting utilities', () => {
  describe('formatBytes', () => {
    it('should format bytes correctly', () => {
      expect(formatBytes(0)).toBe('0 Bytes');
      expect(formatBytes(1024)).toContain('KB');
      expect(formatBytes(1048576)).toContain('MB');
      expect(formatBytes(1073741824)).toContain('GB');
    });

    it('should handle decimal values', () => {
      const result = formatBytes(1536);
      expect(result).toContain('KB');
    });
  });

  describe('formatDuration', () => {
    it('should format milliseconds correctly', () => {
      expect(formatDuration(0)).toContain('ms');
      expect(formatDuration(1000)).toContain('s');
      expect(formatDuration(1500)).toContain('s');
      expect(formatDuration(60000)).toContain('m');
    });
  });

  describe('formatDate', () => {
    it('should format dates correctly', () => {
      const date = new Date('2024-01-01T00:00:00Z');
      const result = formatDate(date);
      expect(result).toBeTruthy();
      expect(typeof result).toBe('string');
    });
  });

  describe('formatRelativeTime', () => {
    it('should format recent times correctly', () => {
      const now = new Date();
      const recent = new Date(now.getTime() - 30 * 1000); // 30 seconds ago
      const result = formatRelativeTime(recent);
      expect(result).toContain('ago');
    });

    it('should format older times correctly', () => {
      const oldDate = new Date('2020-01-01');
      const result = formatRelativeTime(oldDate);
      expect(result).toBeTruthy();
    });
  });
});

