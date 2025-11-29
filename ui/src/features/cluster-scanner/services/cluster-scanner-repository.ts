import { ApiFactory } from '@/core/http/api-factory';
import type {
  Cluster,
  CreateClusterRequest,
  ScanRequest,
  ScanResult,
  ScannedDatabase,
} from '../types';

export class ClusterScannerRepository {
  private client = ApiFactory.getManagerClient();

  // Cluster operations
  async createCluster(request: CreateClusterRequest): Promise<Cluster> {
    return this.client.post<Cluster>('/clusters', request);
  }

  async listClusters(): Promise<Cluster[]> {
    return this.client.get<Cluster[]>('/clusters');
  }

  async getCluster(id: string): Promise<Cluster> {
    return this.client.get<Cluster>(`/clusters/${id}`);
  }

  async deleteCluster(id: string): Promise<void> {
    await this.client.delete(`/clusters/${id}`);
  }

  // Scan operations
  async scanClusters(request: ScanRequest): Promise<ScanResult> {
    return this.client.post<ScanResult>('/clusters/scan', request);
  }

  async getScanResults(clusterId?: string): Promise<ScannedDatabase[]> {
    const url = clusterId
      ? `/clusters/scan/results?cluster_id=${clusterId}`
      : '/clusters/scan/results';
    return this.client.get<ScannedDatabase[]>(url);
  }

  // Discovery operations
  async discoverAvailableClusters(): Promise<AvailableCluster[]> {
    return this.client.get<AvailableCluster[]>('/clusters/discover');
  }
}

export interface AvailableCluster {
  context_name: string;
  cluster_name: string;
  server_url: string;
  type: string;
  provider: string;
  is_current: boolean;
  is_registered: boolean;
  namespace?: string;
  user?: string;
}

export const clusterScannerRepository = new ClusterScannerRepository();

