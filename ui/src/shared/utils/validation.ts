/**
 * Validation Utilities
 */

export function isValidUrl(url: string): boolean {
  try {
    new URL(url);
    return true;
  } catch {
    return false;
  }
}

export function isValidShardName(name: string): boolean {
  return /^[a-z0-9-]+$/.test(name) && name.length >= 3 && name.length <= 50;
}

export function isValidEndpoint(endpoint: string): boolean {
  // Basic validation for PostgreSQL connection string
  return endpoint.startsWith('postgres://') || endpoint.startsWith('postgresql://');
}

export function parseQueryParams(params: string): unknown[] {
  if (!params.trim()) return [];
  try {
    // Try parsing as JSON array first
    return JSON.parse(`[${params}]`);
  } catch {
    // Fallback to comma-separated values
    return params.split(',').map((p) => p.trim());
  }
}

