import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { clusterScannerRepository, type AvailableCluster } from '../services/cluster-scanner-repository';
import type {
  CreateClusterRequest,
  ScanRequest,
} from '../types';

const QUERY_KEYS = {
  clusters: ['clusters'] as const,
  cluster: (id: string) => ['clusters', id] as const,
  scanResults: (clusterId?: string) => ['scan-results', clusterId] as const,
  availableClusters: ['available-clusters'] as const,
};

// Clusters
export function useClusters() {
  return useQuery({
    queryKey: QUERY_KEYS.clusters,
    queryFn: () => clusterScannerRepository.listClusters(),
  });
}

export function useCluster(id: string) {
  return useQuery({
    queryKey: QUERY_KEYS.cluster(id),
    queryFn: () => clusterScannerRepository.getCluster(id),
    enabled: !!id,
  });
}

export function useCreateCluster() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: CreateClusterRequest) =>
      clusterScannerRepository.createCluster(request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.clusters });
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.availableClusters });
    },
  });
}

export function useDeleteCluster() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => clusterScannerRepository.deleteCluster(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.clusters });
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.availableClusters });
    },
  });
}

// Scan operations
export function useScanClusters() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: ScanRequest) =>
      clusterScannerRepository.scanClusters(request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['scan-results'] });
    },
  });
}

export function useScanResults(clusterId?: string) {
  return useQuery({
    queryKey: QUERY_KEYS.scanResults(clusterId),
    queryFn: () => clusterScannerRepository.getScanResults(clusterId),
  });
}

// Discovery operations
export function useAvailableClusters() {
  return useQuery({
    queryKey: QUERY_KEYS.availableClusters,
    queryFn: () => clusterScannerRepository.discoverAvailableClusters(),
    refetchOnWindowFocus: false,
  });
}

