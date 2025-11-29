/**
 * Features Index
 * Central export for all features
 */

// Re-export types for convenience
export type {
  Shard,
  CreateShardRequest,
  PromoteReplicaRequest,
  ShardStatus,
  ShardHealth,
  HealthStatus,
} from './shard/types';

export type {
  QueryRequest,
  QueryResponse,
  ShardInfo,
  ConsistencyLevel,
} from './query/types';

export type {
  ReshardJob,
  SplitRequest,
  MergeRequest,
  ReshardType,
  ReshardStatus,
} from './resharding/types';

export type { HealthResponse } from './health/types';

// Multi-region exports
export * from './multiregion';

// Disaster recovery exports
export * from './disaster-recovery';

// PostgreSQL stats exports
export * from './postgres-stats';

