import { useState, useEffect, useCallback } from 'react';
import { drRepository } from '../services/dr-repository';
import type { DRStatus, DrillResult } from '../types';

export function useDisasterRecovery() {
  const [status, setStatus] = useState<DRStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionInProgress, setActionInProgress] = useState(false);

  const fetchStatus = useCallback(async () => {
    try {
      const data = await drRepository.getStatus();
      setStatus(data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch DR status');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchStatus();
    const interval = setInterval(fetchStatus, 15000);
    return () => clearInterval(interval);
  }, [fetchStatus]);

  const failover = async (targetRegion: string, reason: string) => {
    setActionInProgress(true);
    try {
      await drRepository.failover(targetRegion, reason);
      await fetchStatus();
    } finally {
      setActionInProgress(false);
    }
  };

  const failback = async () => {
    setActionInProgress(true);
    try {
      await drRepository.failback();
      await fetchStatus();
    } finally {
      setActionInProgress(false);
    }
  };

  const runDrill = async (targetRegion: string): Promise<DrillResult> => {
    setActionInProgress(true);
    try {
      const result = await drRepository.runDrill(targetRegion);
      await fetchStatus();
      return result;
    } finally {
      setActionInProgress(false);
    }
  };

  return {
    status,
    loading,
    error,
    actionInProgress,
    refresh: fetchStatus,
    failover,
    failback,
    runDrill,
  };
}

