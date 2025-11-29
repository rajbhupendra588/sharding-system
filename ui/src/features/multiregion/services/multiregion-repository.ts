import type { MultiRegionState, RegionStatus, FailoverRequest, ReplicationLag } from '../types';

const API_BASE = '/api/v1';

export const multiregionRepository = {
  async getState(): Promise<MultiRegionState> {
    const response = await fetch(`${API_BASE}/multiregion/status`);
    if (!response.ok) {
      throw new Error('Failed to fetch multi-region state');
    }
    return response.json();
  },

  async getRegionStatus(region: string): Promise<RegionStatus> {
    const response = await fetch(`${API_BASE}/multiregion/regions/${region}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch region status: ${region}`);
    }
    return response.json();
  },

  async getAllRegions(): Promise<RegionStatus[]> {
    const response = await fetch(`${API_BASE}/multiregion/regions`);
    if (!response.ok) {
      throw new Error('Failed to fetch regions');
    }
    return response.json();
  },

  async initiateFailover(request: FailoverRequest): Promise<void> {
    const response = await fetch(`${API_BASE}/multiregion/failover`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request),
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.message || 'Failed to initiate failover');
    }
  },

  async getReplicationLag(): Promise<ReplicationLag[]> {
    const response = await fetch(`${API_BASE}/multiregion/replication/lag`);
    if (!response.ok) {
      throw new Error('Failed to fetch replication lag');
    }
    return response.json();
  },

  async syncRegion(region: string): Promise<void> {
    const response = await fetch(`${API_BASE}/multiregion/regions/${region}/sync`, {
      method: 'POST',
    });
    if (!response.ok) {
      throw new Error(`Failed to sync region: ${region}`);
    }
  },
};

