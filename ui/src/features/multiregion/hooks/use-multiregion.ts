import { useState, useEffect, useCallback } from 'react';
import { multiregionRepository } from '../services/multiregion-repository';
import type { MultiRegionState, RegionStatus, ReplicationLag } from '../types';

export function useMultiRegion() {
  const [state, setState] = useState<MultiRegionState | null>(null);
  const [regions, setRegions] = useState<RegionStatus[]>([]);
  const [replicationLag, setReplicationLag] = useState<ReplicationLag[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const [stateData, regionsData, lagData] = await Promise.all([
        multiregionRepository.getState(),
        multiregionRepository.getAllRegions(),
        multiregionRepository.getReplicationLag(),
      ]);
      setState(stateData);
      setRegions(regionsData);
      setReplicationLag(lagData);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch data');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 10000); // Refresh every 10s
    return () => clearInterval(interval);
  }, [fetchData]);

  const initiateFailover = async (targetRegion: string, reason: string) => {
    try {
      await multiregionRepository.initiateFailover({ target_region: targetRegion, reason });
      await fetchData();
    } catch (err) {
      throw err;
    }
  };

  const syncRegion = async (region: string) => {
    try {
      await multiregionRepository.syncRegion(region);
      await fetchData();
    } catch (err) {
      throw err;
    }
  };

  return {
    state,
    regions,
    replicationLag,
    loading,
    error,
    refresh: fetchData,
    initiateFailover,
    syncRegion,
  };
}

