/**
 * Application Constants
 * Centralized configuration and constants
 */

export const API_CONFIG = {
  // Use relative URLs in K8s - nginx will proxy them
  MANAGER_BASE_URL: import.meta.env.VITE_MANAGER_URL || '',
  ROUTER_BASE_URL: import.meta.env.VITE_ROUTER_URL || '',
  MANAGER_API_PREFIX: '/api/v1',
  ROUTER_API_PREFIX: '/v1',
  TIMEOUT: 30000,
} as const;

export const REFRESH_INTERVALS = {
  SHARDS: 10000, // 10 seconds
  HEALTH: 5000, // 5 seconds
  RESHARD_JOB_ACTIVE: 5000, // 5 seconds for active jobs
  RESHARD_JOB_INACTIVE: false, // No auto-refresh for completed/failed
} as const;

export const STORAGE_KEYS = {
  AUTH_TOKEN: 'auth_token',
  MANAGER_URL: 'manager_url',
  ROUTER_URL: 'router_url',
  REFRESH_INTERVAL: 'refresh_interval',
} as const;

export const ROUTES = {
  DASHBOARD: '/dashboard',
  SHARDS: '/shards',
  SHARD_DETAIL: (id: string) => `/shards/${id}`,
  QUERY: '/query',
  RESHARDING: '/resharding',
  RESHARD_JOB_DETAIL: (id: string) => `/resharding/jobs/${id}`,
  HEALTH: '/health',
  METRICS: '/metrics',
  SETTINGS: '/settings',
} as const;

export const SHARD_STATUS = {
  ACTIVE: 'active',
  MIGRATING: 'migrating',
  READONLY: 'readonly',
  INACTIVE: 'inactive',
} as const;

export const RESHARD_STATUS = {
  PENDING: 'pending',
  PRECOPY: 'precopy',
  DELTASYNC: 'deltasync',
  CUTOVER: 'cutover',
  VALIDATION: 'validation',
  COMPLETED: 'completed',
  FAILED: 'failed',
} as const;

export const HEALTH_STATUS = {
  HEALTHY: 'healthy',
  DEGRADED: 'degraded',
  UNHEALTHY: 'unhealthy',
} as const;

export const CONSISTENCY_LEVELS = {
  STRONG: 'strong',
  EVENTUAL: 'eventual',
} as const;

export const RESHARD_TYPES = {
  SPLIT: 'split',
  MERGE: 'merge',
} as const;

export const DEFAULT_VALUES = {
  VNODE_COUNT: 256,
  MIN_VNODE_COUNT: 1,
  MAX_VNODE_COUNT: 1024,
} as const;

