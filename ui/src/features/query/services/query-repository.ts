/**
 * Query Repository
 * Data access layer for query operations
 */

import { ApiFactory } from '@/core/http/api-factory';
import type { QueryRequest, QueryResponse, ShardInfo } from '../types';

export class QueryRepository {
  private client = ApiFactory.getRouterClient();

  async execute(request: QueryRequest): Promise<QueryResponse> {
    return this.client.post<QueryResponse>('/execute', request);
  }

  async getShardForKey(key: string): Promise<ShardInfo> {
    return this.client.get<ShardInfo>(`/shard-for-key?key=${encodeURIComponent(key)}`);
  }
}

export const queryRepository = new QueryRepository();

