export interface Cluster {
  id: string;
  name: string;
  type: 'cloud' | 'onprem';
  provider?: string;
  kubeconfig?: string;
  context?: string;
  endpoint?: string;
  credentials?: Record<string, string>;
  status: 'active' | 'inactive' | 'error';
  last_scan?: string;
  created_at: string;
  updated_at: string;
  metadata?: Record<string, string>;
}

export interface CreateClusterRequest {
  name: string;
  type: 'cloud' | 'onprem';
  provider?: string;
  kubeconfig?: string;
  context?: string;
  endpoint?: string;
  credentials?: Record<string, string>;
  metadata?: Record<string, string>;
}

export interface ScannedDatabase {
  id: string;
  cluster_id: string;
  cluster_name: string;
  namespace: string;
  app_name: string;
  app_type: string;
  database_name: string;
  database_type: string;
  host: string;
  port: number;
  database: string;
  username?: string;
  status: 'discovered' | 'scanned' | 'error';
  scan_error?: string;
  scan_results?: DatabaseScanResults;
  discovered_at: string;
  last_scanned_at?: string;
  labels?: Record<string, string>;
  annotations?: Record<string, string>;
}

export interface DatabaseScanResults {
  version?: string;
  size?: number;
  table_count?: number;
  table_names?: string[];
  index_count?: number;
  connection_count?: number;
  max_connections?: number;
  is_replica?: boolean;
  replication_lag?: number;
  uptime?: number;
  table_stats?: TableStat[];
  index_stats?: IndexStat[];
  health_status: 'healthy' | 'degraded' | 'unhealthy';
  metadata?: Record<string, unknown>;
}

export interface TableStat {
  name: string;
  row_count: number;
  size: number;
  index_size: number;
  total_size: number;
  last_vacuum?: string;
  last_analyze?: string;
}

export interface IndexStat {
  name: string;
  table_name: string;
  size: number;
  scans: number;
  tuples_read: number;
  tuples_fetched: number;
}

export interface ScanRequest {
  cluster_ids?: string[];
  deep_scan?: boolean;
}

export interface ScanResult {
  id: string;
  cluster_id: string;
  status: 'running' | 'completed' | 'failed';
  databases_found: number;
  databases_scanned: number;
  databases_failed: number;
  started_at: string;
  completed_at?: string;
  error?: string;
  results?: ScannedDatabase[];
}

