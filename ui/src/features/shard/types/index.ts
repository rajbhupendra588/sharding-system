/**
 * Shard Feature Types
 * Types specific to shard management
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
  client_app_id?: string;
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

