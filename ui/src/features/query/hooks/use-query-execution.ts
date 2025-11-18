/**
 * useQueryExecution Hook
 * Business logic hook for query execution
 */

import { useMutation } from '@tanstack/react-query';
import { queryRepository } from '../services/query-repository';
import type { QueryRequest } from '../types';
import { toast } from 'react-hot-toast';

export function useExecuteQuery() {
  return useMutation({
    mutationFn: (request: QueryRequest) => queryRepository.execute(request),
    onSuccess: () => {
      toast.success('Query executed successfully');
    },
    onError: (error: { message: string }) => {
      toast.error(`Query failed: ${error.message}`);
    },
  });
}

export function useGetShardForKey() {
  return useMutation({
    mutationFn: (key: string) => queryRepository.getShardForKey(key),
    onError: (error: { message: string }) => {
      toast.error(`Failed to get shard: ${error.message}`);
    },
  });
}

