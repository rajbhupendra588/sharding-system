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
  updateShardStatus: (id: string, status: string) => shardRepository.updateStatus(id, status),

  // Client App operations
  deleteClientApp: async (id: string) => {
    const response = await fetch(`/api/v1/client-apps/${id}`, { method: 'DELETE' });
    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: 'Failed to delete client app' }));
      throw new Error(error.message || 'Failed to delete client app');
    }
  },

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
