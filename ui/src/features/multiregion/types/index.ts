export interface RegionConfig {
  name: string;
  endpoint: string;
  priority: number;
  weight: number;
  health_check_path: string;
  is_local: boolean;
  metadata?: Record<string, string>;
}

export interface RegionStatus {
  name: string;
  endpoint: string;
  is_healthy: boolean;
  is_primary: boolean;
  latency: number; // in ms
  last_check: string;
  error_count: number;
  active_connections: number;
}

export interface MultiRegionState {
  local_region: string;
  primary_region: string;
  regions: RegionStatus[];
  failover_enabled: boolean;
  sync_interval: string;
}

export interface FailoverRequest {
  target_region: string;
  reason: string;
}

export interface ReplicationLag {
  region: string;
  lag_ms: number;
  last_sync: string;
  status: 'synced' | 'lagging' | 'stale';
}

