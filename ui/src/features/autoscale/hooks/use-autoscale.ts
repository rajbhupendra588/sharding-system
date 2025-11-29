import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { autoscaleRepository } from '../services/autoscale-repository';
import { metricsRepository } from '../services/metrics-repository';
import type { Thresholds } from '../types';
import { toast } from 'react-hot-toast';

export function useAutoscaleStatus() {
  return useQuery({
    queryKey: ['autoscale', 'status'],
    queryFn: () => autoscaleRepository.getStatus(),
    refetchInterval: 5000,
  });
}

export function useHotShards() {
  return useQuery({
    queryKey: ['autoscale', 'hot-shards'],
    queryFn: () => autoscaleRepository.getHotShards(),
    refetchInterval: 10000,
  });
}

export function useColdShards() {
  return useQuery({
    queryKey: ['autoscale', 'cold-shards'],
    queryFn: () => autoscaleRepository.getColdShards(),
    refetchInterval: 10000,
  });
}

export function useThresholds() {
  return useQuery({
    queryKey: ['autoscale', 'thresholds'],
    queryFn: () => autoscaleRepository.getThresholds(),
  });
}

export function useEnableAutoscale() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => autoscaleRepository.enable(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['autoscale', 'status'] });
      toast.success('Auto-scaling enabled');
    },
    onError: (error: any) => {
      toast.error(`Failed to enable auto-scaling: ${error.message}`);
    },
  });
}

export function useDisableAutoscale() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => autoscaleRepository.disable(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['autoscale', 'status'] });
      toast.success('Auto-scaling disabled');
    },
    onError: (error: any) => {
      toast.error(`Failed to disable auto-scaling: ${error.message}`);
    },
  });
}

export function useUpdateThresholds() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (thresholds: Thresholds) => autoscaleRepository.updateThresholds(thresholds),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['autoscale', 'thresholds'] });
      toast.success('Thresholds updated');
    },
    onError: (error: any) => {
      toast.error(`Failed to update thresholds: ${error.message}`);
    },
  });
}

export function useShardMetrics(shardID: string) {
  return useQuery({
    queryKey: ['metrics', 'shard', shardID],
    queryFn: () => metricsRepository.getShardMetrics(shardID),
    enabled: !!shardID,
    refetchInterval: 5000,
  });
}

export function useAllMetrics() {
  return useQuery({
    queryKey: ['metrics', 'all'],
    queryFn: () => metricsRepository.getAllMetrics(),
    refetchInterval: 5000,
  });
}

