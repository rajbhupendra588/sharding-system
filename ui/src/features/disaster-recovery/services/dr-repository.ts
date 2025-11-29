import type { DRStatus, DrillResult, DRAction } from '../types';

const API_BASE = '/api/v1';

export const drRepository = {
  async getStatus(): Promise<DRStatus> {
    const response = await fetch(`${API_BASE}/disaster-recovery`);
    if (!response.ok) {
      throw new Error('Failed to fetch DR status');
    }
    return response.json();
  },

  async performAction(action: DRAction): Promise<DrillResult | void> {
    const response = await fetch(`${API_BASE}/disaster-recovery`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(action),
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.message || `Failed to perform ${action.action}`);
    }
    if (action.action === 'drill') {
      return response.json();
    }
  },

  async failover(targetRegion: string, reason: string): Promise<void> {
    await this.performAction({ action: 'failover', target: targetRegion, reason });
  },

  async failback(): Promise<void> {
    await this.performAction({ action: 'failback' });
  },

  async runDrill(targetRegion: string): Promise<DrillResult> {
    const result = await this.performAction({ action: 'drill', target: targetRegion });
    return result as DrillResult;
  },
};

