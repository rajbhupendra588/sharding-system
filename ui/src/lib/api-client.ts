/**
 * Legacy API Client - Deprecated
 * Use feature-specific repositories instead
 * @deprecated Use feature repositories (e.g., shardRepository, queryRepository)
 */

import { shardRepository } from '@/features/shard';
import { queryRepository } from '@/features/query';
import { reshardRepository } from '@/features/resharding';
import { healthRepository } from '@/features/health';
import type {
  CreateShardRequest,
  QueryRequest,
  SplitRequest,
  MergeRequest,
  PromoteReplicaRequest,
} from '@/features';

// Re-export for backward compatibility
export const apiClient = {
  // Shard operations
  createShard: (request: CreateShardRequest) => shardRepository.create(request),
  listShards: () => shardRepository.findAll(),
  getShard: (id: string) => shardRepository.findById(id),
  deleteShard: (id: string) => shardRepository.delete(id),
  promoteReplica: (shardId: string, request: PromoteReplicaRequest) =>
    shardRepository.promoteReplica(shardId, request),

  // Query operations
  executeQuery: (request: QueryRequest) => queryRepository.execute(request),
  getShardForKey: (key: string) => queryRepository.getShardForKey(key),

  // Resharding operations
  splitShard: (request: SplitRequest) => reshardRepository.split(request),
  mergeShards: (request: MergeRequest) => reshardRepository.merge(request),
  getReshardJob: (jobId: string) => reshardRepository.getJob(jobId),

  // Health operations
  getManagerHealth: () => healthRepository.getManagerHealth(),
  getRouterHealth: () => healthRepository.getRouterHealth(),
};
