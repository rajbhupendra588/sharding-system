import { ApiFactory } from '@/core/http/api-factory';
import type { ShardMetrics } from '../types';

export class MetricsRepository {
  private client = ApiFactory.getManagerClient();

  async getShardMetrics(shardID: string): Promise<ShardMetrics> {
    return this.client.get<ShardMetrics>(`/metrics/shard/${shardID}`);
  }

  async getAllMetrics(): Promise<Record<string, ShardMetrics>> {
    return this.client.get<Record<string, ShardMetrics>>('/metrics/shard');
  }
}

export const metricsRepository = new MetricsRepository();

