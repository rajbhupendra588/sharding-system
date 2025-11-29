import { useState, useEffect, useCallback } from 'react';
import { postgresStatsRepository } from '../services/postgres-stats-repository';
import type { PostgresStats } from '../types';

export function usePostgresStats(databaseId?: string) {
  const [stats, setStats] = useState<PostgresStats | null>(null);
  const [allStats, setAllStats] = useState<Record<string, PostgresStats>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchStats = useCallback(async () => {
    try {
      if (databaseId) {
        const data = await postgresStatsRepository.getStats(databaseId);
        setStats(data);
      } else {
        const data = await postgresStatsRepository.getAllStats();
        setAllStats(data);
      }
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch stats');
    } finally {
      setLoading(false);
    }
  }, [databaseId]);

  useEffect(() => {
    fetchStats();
    const interval = setInterval(fetchStats, 30000); // Refresh every 30s
    return () => clearInterval(interval);
  }, [fetchStats]);

  const runVacuum = async (dbId: string) => {
    await postgresStatsRepository.runVacuum(dbId);
    await fetchStats();
  };

  const runAnalyze = async (dbId: string) => {
    await postgresStatsRepository.runAnalyze(dbId);
    await fetchStats();
  };

  return {
    stats,
    allStats,
    loading,
    error,
    refresh: fetchStats,
    runVacuum,
    runAnalyze,
  };
}

