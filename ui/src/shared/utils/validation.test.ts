import { describe, it, expect } from 'vitest';
import {
  isValidUrl,
  isValidShardName,
  isValidEndpoint,
  parseQueryParams,
} from './validation';

describe('validation utilities', () => {
  describe('isValidUrl', () => {
    it('should return true for valid HTTP URLs', () => {
      expect(isValidUrl('http://example.com')).toBe(true);
      expect(isValidUrl('http://example.com/path')).toBe(true);
      expect(isValidUrl('http://example.com:8080/path')).toBe(true);
    });

    it('should return true for valid HTTPS URLs', () => {
      expect(isValidUrl('https://example.com')).toBe(true);
      expect(isValidUrl('https://example.com/path')).toBe(true);
    });

    it('should return false for invalid URLs', () => {
      expect(isValidUrl('not-a-url')).toBe(false);
      expect(isValidUrl('http://')).toBe(false);
      expect(isValidUrl('example.com')).toBe(false);
      expect(isValidUrl('')).toBe(false);
    });

    it('should handle URLs with query parameters', () => {
      expect(isValidUrl('http://example.com?param=value')).toBe(true);
      expect(isValidUrl('https://example.com?foo=bar&baz=qux')).toBe(true);
    });

    it('should handle URLs with fragments', () => {
      expect(isValidUrl('http://example.com#section')).toBe(true);
    });
  });

  describe('isValidShardName', () => {
    it('should return true for valid shard names', () => {
      expect(isValidShardName('shard-1')).toBe(true);
      expect(isValidShardName('shard01')).toBe(true);
      expect(isValidShardName('my-shard')).toBe(true);
      expect(isValidShardName('abc')).toBe(true); // minimum length
      expect(isValidShardName('a'.repeat(50))).toBe(true); // maximum length
    });

    it('should return false for shard names with uppercase letters', () => {
      expect(isValidShardName('Shard-1')).toBe(false);
      expect(isValidShardName('SHARD-1')).toBe(false);
    });

    it('should return false for shard names with special characters', () => {
      expect(isValidShardName('shard_1')).toBe(false);
      expect(isValidShardName('shard.1')).toBe(false);
      expect(isValidShardName('shard@1')).toBe(false);
      expect(isValidShardName('shard 1')).toBe(false);
    });

    it('should return false for shard names that are too short', () => {
      expect(isValidShardName('ab')).toBe(false);
      expect(isValidShardName('a')).toBe(false);
      expect(isValidShardName('')).toBe(false);
    });

    it('should return false for shard names that are too long', () => {
      expect(isValidShardName('a'.repeat(51))).toBe(false);
      expect(isValidShardName('a'.repeat(100))).toBe(false);
    });

    it('should allow hyphens in shard names', () => {
      expect(isValidShardName('shard-1')).toBe(true);
      expect(isValidShardName('my-shard-name')).toBe(true);
      expect(isValidShardName('shard--double')).toBe(true);
    });
  });

  describe('isValidEndpoint', () => {
    it('should return true for postgres:// URLs', () => {
      expect(isValidEndpoint('postgres://user:pass@host:5432/db')).toBe(true);
      expect(isValidEndpoint('postgres://localhost/dbname')).toBe(true);
    });

    it('should return true for postgresql:// URLs', () => {
      expect(isValidEndpoint('postgresql://user:pass@host:5432/db')).toBe(true);
      expect(isValidEndpoint('postgresql://localhost/dbname')).toBe(true);
    });

    it('should return false for other protocols', () => {
      expect(isValidEndpoint('http://example.com')).toBe(false);
      expect(isValidEndpoint('https://example.com')).toBe(false);
      expect(isValidEndpoint('mysql://localhost/db')).toBe(false);
    });

    it('should return false for invalid strings', () => {
      expect(isValidEndpoint('not-an-endpoint')).toBe(false);
      expect(isValidEndpoint('')).toBe(false);
    });

    it('should be case-sensitive for protocol', () => {
      expect(isValidEndpoint('Postgres://localhost/db')).toBe(false);
      expect(isValidEndpoint('POSTGRES://localhost/db')).toBe(false);
    });
  });

  describe('parseQueryParams', () => {
    it('should parse JSON array format', () => {
      // parseQueryParams parses numbers as numbers when using JSON.parse
      expect(parseQueryParams('1,2,3')).toEqual([1, 2, 3]);
      expect(parseQueryParams('"a","b","c"')).toEqual(['a', 'b', 'c']);
    });

    it('should parse comma-separated values', () => {
      expect(parseQueryParams('value1,value2,value3')).toEqual(['value1', 'value2', 'value3']);
    });

    it('should trim whitespace from values', () => {
      expect(parseQueryParams(' value1 , value2 , value3 ')).toEqual(['value1', 'value2', 'value3']);
    });

    it('should return empty array for empty string', () => {
      expect(parseQueryParams('')).toEqual([]);
      expect(parseQueryParams('   ')).toEqual([]);
    });

    it('should handle single value', () => {
      expect(parseQueryParams('single')).toEqual(['single']);
    });

    it('should handle JSON array with numbers', () => {
      const result = parseQueryParams('1,2,3');
      // parseQueryParams returns numbers when parsing numeric values
      expect(result).toEqual([1, 2, 3]);
    });

    it('should handle mixed JSON and comma-separated format', () => {
      // The function tries JSON first, then falls back to comma-separated
      const result = parseQueryParams('a,b,c');
      expect(result).toEqual(['a', 'b', 'c']);
    });
  });
});

