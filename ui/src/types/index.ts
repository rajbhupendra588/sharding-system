/**
 * Type definitions for the Sharding System UI
 * All types match the backend API models
 */

export interface Shard {
  id: string;
  name: string;
  hash_range_start?: number;
  hash_range_end?: number;
  primary_endpoint: string;
  replicas: string[];
  status: ShardStatus;
  version: number;
  created_at: string;
  updated_at: string;
  vnodes?: VNode[];
}

export type ShardStatus = 'active' | 'migrating' | 'readonly' | 'inactive';

export interface VNode {
  id: number;
  shard_id: string;
  hash: number;
}

export interface ShardCatalog {
  version: number;
  shards: Shard[];
  updated_at: string;
}

export interface CreateShardRequest {
  name: string;
  primary_endpoint: string;
  replicas: string[];
  vnode_count: number;
}

export interface QueryRequest {
  shard_key: string;
  query: string;
  params: unknown[];
  consistency: ConsistencyLevel;
  options?: Record<string, unknown>;
}

export type ConsistencyLevel = 'strong' | 'eventual';

export interface QueryResponse {
  shard_id: string;
  rows: Record<string, unknown>[];
  row_count: number;
  latency_ms: number;
}

export interface ShardInfo {
  shard_id: string;
  shard_name?: string;
  hash_value?: number;
}

export interface ReshardJob {
  id: string;
  type: ReshardType;
  source_shards: string[];
  target_shards: string[];
  status: ReshardStatus;
  progress: number; // 0.0 to 1.0
  started_at: string;
  completed_at?: string;
  error_message?: string;
  keys_migrated: number;
  total_keys: number;
}

export type ReshardType = 'split' | 'merge';
export type ReshardStatus = 'pending' | 'precopy' | 'deltasync' | 'cutover' | 'validation' | 'completed' | 'failed';

export interface SplitRequest {
  source_shard_id: string;
  target_shards: CreateShardRequest[];
  split_point?: number;
}

export interface MergeRequest {
  source_shard_ids: string[];
  target_shard: CreateShardRequest;
}

export interface PromoteReplicaRequest {
  replica_endpoint: string;
}

export interface ShardHealth {
  shard_id: string;
  status: HealthStatus;
  replication_lag: number; // seconds
  last_check: string;
  primary_up: boolean;
  replicas_up: string[];
  replicas_down: string[];
}

export type HealthStatus = 'healthy' | 'degraded' | 'unhealthy';

export interface HealthResponse {
  status: string;
  version?: string;
  uptime_seconds?: number;
}

export interface ApiError {
  message: string;
  code?: string;
  details?: unknown;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
}

