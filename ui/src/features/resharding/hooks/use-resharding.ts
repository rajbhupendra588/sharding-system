/**
 * useResharding Hook
 * Business logic hook for resharding operations
 */

import { useQuery, useMutation } from '@tanstack/react-query';
import { reshardRepository } from '../services/reshard-repository';
import { REFRESH_INTERVALS, RESHARD_STATUS } from '@/core/config';
import type { SplitRequest, MergeRequest } from '../types';
import { toast } from 'react-hot-toast';

const QUERY_KEYS = {
  all: ['reshard-jobs'] as const,
  details: () => [...QUERY_KEYS.all, 'detail'] as const,
  detail: (id: string) => [...QUERY_KEYS.details(), id] as const,
};

export function useReshardJob(jobId: string | undefined) {
  return useQuery({
    queryKey: QUERY_KEYS.detail(jobId!),
    queryFn: () => reshardRepository.getJob(jobId!),
    enabled: !!jobId,
    refetchInterval: (query) => {
      // Auto-refresh if job is still in progress
      const job = query.state.data;
      if (
        job?.status &&
        [
          RESHARD_STATUS.PENDING,
          RESHARD_STATUS.PRECOPY,
          RESHARD_STATUS.DELTASYNC,
          RESHARD_STATUS.CUTOVER,
          RESHARD_STATUS.VALIDATION,
        ].includes(job.status as any)
      ) {
        return REFRESH_INTERVALS.RESHARD_JOB_ACTIVE;
      }
      return false;
    },
  });
}

export function useSplitShard() {
  return useMutation({
    mutationFn: (request: SplitRequest) => reshardRepository.split(request),
    onSuccess: () => {
      toast.success('Split operation started');
    },
    onError: (error: { message: string }) => {
      toast.error(`Failed to start split: ${error.message}`);
    },
  });
}

export function useMergeShards() {
  return useMutation({
    mutationFn: (request: MergeRequest) => reshardRepository.merge(request),
    onSuccess: () => {
      toast.success('Merge operation started');
    },
    onError: (error: { message: string }) => {
      toast.error(`Failed to start merge: ${error.message}`);
    },
  });
}

