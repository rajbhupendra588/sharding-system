/**
 * Query Feature Types
 */

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

