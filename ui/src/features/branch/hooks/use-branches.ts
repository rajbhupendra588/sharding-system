import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { branchRepository } from '../services/branch-repository';
import type { CreateBranchRequest } from '../types';
import { toast } from 'react-hot-toast';

export function useBranches(dbName: string) {
  return useQuery({
    queryKey: ['branches', dbName],
    queryFn: () => branchRepository.listBranches(dbName),
    enabled: !!dbName,
    refetchInterval: 10000,
  });
}

export function useBranch(branchID: string) {
  return useQuery({
    queryKey: ['branches', branchID],
    queryFn: () => branchRepository.getBranch(branchID),
    enabled: !!branchID,
    refetchInterval: 10000,
  });
}

export function useCreateBranch() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ dbName, request }: { dbName: string; request: CreateBranchRequest }) =>
      branchRepository.createBranch(dbName, request),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['branches', variables.dbName] });
      toast.success('Branch creation initiated');
    },
    onError: (error: any) => {
      toast.error(`Failed to create branch: ${error.message}`);
    },
  });
}

export function useDeleteBranch() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (branchID: string) => branchRepository.deleteBranch(branchID),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['branches'] });
      toast.success('Branch deletion initiated');
    },
    onError: (error: any) => {
      toast.error(`Failed to delete branch: ${error.message}`);
    },
  });
}

export function useMergeBranch() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (branchID: string) => branchRepository.mergeBranch(branchID),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['branches'] });
      toast.success('Branch merge initiated');
    },
    onError: (error: any) => {
      toast.error(`Failed to merge branch: ${error.message}`);
    },
  });
}

