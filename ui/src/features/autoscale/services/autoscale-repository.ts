import { ApiFactory } from '@/core/http/api-factory';
import type { AutoscaleStatus, Thresholds } from '../types';

export class AutoscaleRepository {
  private client = ApiFactory.getManagerClient();

  async getStatus(): Promise<AutoscaleStatus> {
    return this.client.get<AutoscaleStatus>('/autoscale/status');
  }

  async enable(): Promise<void> {
    await this.client.post('/autoscale/enable');
  }

  async disable(): Promise<void> {
    await this.client.post('/autoscale/disable');
  }

  async getHotShards(): Promise<string[]> {
    const response = await this.client.get<{ shards: string[] }>('/autoscale/hot-shards');
    return response.shards;
  }

  async getColdShards(): Promise<string[]> {
    const response = await this.client.get<{ shards: string[] }>('/autoscale/cold-shards');
    return response.shards;
  }

  async getThresholds(): Promise<Thresholds> {
    return this.client.get<Thresholds>('/autoscale/thresholds');
  }

  async updateThresholds(thresholds: Thresholds): Promise<Thresholds> {
    return this.client.put<Thresholds>('/autoscale/thresholds', thresholds);
  }
}

export const autoscaleRepository = new AutoscaleRepository();

