/**
 * useHealth Hook
 * Business logic hook for health monitoring
 */

import { useQuery } from '@tanstack/react-query';
import { healthRepository } from '../services/health-repository';
import { REFRESH_INTERVALS } from '@/core/config';

const QUERY_KEYS = {
  all: ['health'] as const,
  manager: () => [...QUERY_KEYS.all, 'manager'] as const,
  router: () => [...QUERY_KEYS.all, 'router'] as const,
};

export function useManagerHealth() {
  return useQuery({
    queryKey: QUERY_KEYS.manager(),
    queryFn: () => healthRepository.getManagerHealth(),
    refetchInterval: REFRESH_INTERVALS.HEALTH,
  });
}

export function useRouterHealth() {
  return useQuery({
    queryKey: QUERY_KEYS.router(),
    queryFn: () => healthRepository.getRouterHealth(),
    refetchInterval: REFRESH_INTERVALS.HEALTH,
  });
}

export function useSystemHealth() {
  const managerHealth = useManagerHealth();
  const routerHealth = useRouterHealth();

  const isHealthy =
    managerHealth.data?.status === 'healthy' &&
    routerHealth.data?.status === 'healthy';

  return {
    managerHealth,
    routerHealth,
    isHealthy,
    isLoading: managerHealth.isLoading || routerHealth.isLoading,
  };
}

