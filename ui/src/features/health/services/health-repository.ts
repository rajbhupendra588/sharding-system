/**
 * Health Repository
 * Data access layer for health checks
 */

import { ApiFactory } from '@/core/http/api-factory';
import type { HealthResponse } from '../types';

export class HealthRepository {
  private managerClient = ApiFactory.getManagerClient();
  private routerClient = ApiFactory.getRouterClient();

  async getManagerHealth(): Promise<HealthResponse> {
    // baseURL already includes /api/v1, so just use /health
    return await this.managerClient.get<HealthResponse>('/health');
  }

  async getRouterHealth(): Promise<HealthResponse> {
    // baseURL already includes /v1, so just use /health
    return await this.routerClient.get<HealthResponse>('/health');
  }
}

export const healthRepository = new HealthRepository();

