/**
 * Reshard Repository
 * Data access layer for resharding operations
 */

import { ApiFactory } from '@/core/http/api-factory';
import type { ReshardJob, SplitRequest, MergeRequest } from '../types';

export class ReshardRepository {
  private client = ApiFactory.getManagerClient();

  async split(request: SplitRequest): Promise<ReshardJob> {
    return this.client.post<ReshardJob>('/reshard/split', request);
  }

  async merge(request: MergeRequest): Promise<ReshardJob> {
    return this.client.post<ReshardJob>('/reshard/merge', request);
  }

  async getJob(jobId: string): Promise<ReshardJob> {
    return this.client.get<ReshardJob>(`/reshard/jobs/${jobId}`);
  }
}

export const reshardRepository = new ReshardRepository();

