import type { PostgresStats } from '../types';

const API_BASE = '/api/v1';

export const postgresStatsRepository = {
  async getStats(databaseId: string): Promise<PostgresStats> {
    const response = await fetch(`${API_BASE}/databases/${databaseId}/stats`);
    if (!response.ok) {
      throw new Error(`Failed to fetch stats for database: ${databaseId}`);
    }
    return response.json();
  },

  async getAllStats(): Promise<Record<string, PostgresStats>> {
    const response = await fetch(`${API_BASE}/databases/stats`);
    if (!response.ok) {
      throw new Error('Failed to fetch database stats');
    }
    return response.json();
  },

  async getShardStats(shardId: string): Promise<PostgresStats> {
    const response = await fetch(`${API_BASE}/shards/${shardId}/stats`);
    if (!response.ok) {
      throw new Error(`Failed to fetch stats for shard: ${shardId}`);
    }
    return response.json();
  },

  async runVacuum(databaseId: string): Promise<void> {
    const response = await fetch(`${API_BASE}/databases/${databaseId}/vacuum`, {
      method: 'POST',
    });
    if (!response.ok) {
      throw new Error('Failed to run vacuum');
    }
  },

  async runAnalyze(databaseId: string): Promise<void> {
    const response = await fetch(`${API_BASE}/databases/${databaseId}/analyze`, {
      method: 'POST',
    });
    if (!response.ok) {
      throw new Error('Failed to run analyze');
    }
  },
};

