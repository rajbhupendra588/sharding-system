import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { databaseRepository } from '../services/database-repository';
import type { CreateDatabaseRequest } from '../types';

export function useDatabases() {
  return useQuery({
    queryKey: ['databases'],
    queryFn: () => databaseRepository.findAll(),
    refetchInterval: 10000, // Refetch every 10 seconds
  });
}

export function useDatabase(id: string) {
  return useQuery({
    queryKey: ['databases', id],
    queryFn: () => databaseRepository.findById(id),
    enabled: !!id,
    refetchInterval: 10000,
  });
}

export function useDatabaseStatus(id: string) {
  return useQuery({
    queryKey: ['databases', id, 'status'],
    queryFn: () => databaseRepository.getStatus(id),
    enabled: !!id,
    refetchInterval: 5000, // More frequent updates for status
  });
}

export function useDatabaseTemplates() {
  return useQuery({
    queryKey: ['database-templates'],
    queryFn: () => databaseRepository.getTemplates(),
    staleTime: 5 * 60 * 1000, // Templates don't change often
  });
}

export function useCreateDatabase() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: CreateDatabaseRequest) => databaseRepository.create(request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['databases'] });
    },
  });
}

export function useBackups(databaseId: string) {
  return useQuery({
    queryKey: ['databases', databaseId, 'backups'],
    queryFn: () => databaseRepository.listBackups(databaseId),
    enabled: !!databaseId,
    refetchInterval: 30000, // Refetch every 30 seconds
  });
}

export function useCreateBackup() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ databaseId, type }: { databaseId: string; type?: 'full' | 'incremental' }) =>
      databaseRepository.createBackup(databaseId, type),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['databases', variables.databaseId, 'backups'] });
    },
  });
}

export function useFailoverStatus() {
  return useQuery({
    queryKey: ['failover-status'],
    queryFn: () => databaseRepository.getFailoverStatus(),
    refetchInterval: 10000,
  });
}

export function useFailoverHistory(shardId?: string) {
  return useQuery({
    queryKey: ['failover-history', shardId],
    queryFn: () => databaseRepository.getFailoverHistory(shardId),
    refetchInterval: 30000,
  });
}

