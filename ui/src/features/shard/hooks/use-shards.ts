/**
 * useShards Hook
 * Business logic hook for shard management
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { shardRepository } from '../services/shard-repository';
import { REFRESH_INTERVALS } from '@/core/config';
import type { CreateShardRequest, PromoteReplicaRequest } from '../types';
import { toast } from 'react-hot-toast';

const QUERY_KEYS = {
  all: ['shards'] as const,
  lists: () => [...QUERY_KEYS.all, 'list'] as const,
  list: (filters?: string) => [...QUERY_KEYS.lists(), { filters }] as const,
  details: () => [...QUERY_KEYS.all, 'detail'] as const,
  detail: (id: string) => [...QUERY_KEYS.details(), id] as const,
};

export function useShards() {
  return useQuery({
    queryKey: QUERY_KEYS.lists(),
    queryFn: () => shardRepository.findAll(),
    refetchInterval: REFRESH_INTERVALS.SHARDS,
    staleTime: 5000,
  });
}

export function useShard(id: string | undefined) {
  return useQuery({
    queryKey: QUERY_KEYS.detail(id!),
    queryFn: () => shardRepository.findById(id!),
    enabled: !!id,
    refetchInterval: REFRESH_INTERVALS.SHARDS,
  });
}

export function useCreateShard() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateShardRequest) => shardRepository.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.all });
      toast.success('Shard created successfully');
    },
    onError: (error: { message: string }) => {
      toast.error(`Failed to create shard: ${error.message}`);
    },
  });
}

export function useDeleteShard() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => shardRepository.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.all });
      toast.success('Shard deleted successfully');
    },
    onError: (error: { message: string }) => {
      toast.error(`Failed to delete shard: ${error.message}`);
    },
  });
}

export function usePromoteReplica(shardId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: PromoteReplicaRequest) =>
      shardRepository.promoteReplica(shardId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.detail(shardId) });
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.all });
      toast.success('Replica promoted successfully');
    },
    onError: (error: { message: string }) => {
      toast.error(`Failed to promote replica: ${error.message}`);
    },
  });
}

