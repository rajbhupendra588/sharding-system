/**
 * Shard Repository
 * Data access layer for shard operations
 */

import { ApiFactory } from '@/core/http/api-factory';
import type {
  Shard,
  CreateShardRequest,
  PromoteReplicaRequest,
} from '../types';

export class ShardRepository {
  private client = ApiFactory.getManagerClient();

  async create(request: CreateShardRequest): Promise<Shard> {
    return this.client.post<Shard>('/shards', request);
  }

  async findAll(): Promise<Shard[]> {
    return this.client.get<Shard[]>('/shards');
  }

  async findById(id: string): Promise<Shard> {
    return this.client.get<Shard>(`/shards/${id}`);
  }

  async delete(id: string): Promise<void> {
    return this.client.delete(`/shards/${id}`);
  }

  async promoteReplica(shardId: string, request: PromoteReplicaRequest): Promise<void> {
    return this.client.post(`/shards/${shardId}/promote`, request);
  }

  async updateStatus(id: string, status: string): Promise<void> {
    return this.client.put(`/shards/${id}/status`, { status });
  }
}

export const shardRepository = new ShardRepository();

