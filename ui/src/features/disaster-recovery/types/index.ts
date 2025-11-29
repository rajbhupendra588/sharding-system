export interface DRStatus {
  current_region: string;
  primary_region: string;
  is_failed_over: boolean;
  rpo: string;
  rto: string;
  auto_failover: boolean;
  region_statuses: RegionHealthStatus[];
  failover_history: FailoverEvent[];
}

export interface RegionHealthStatus {
  region: string;
  is_healthy: boolean;
  last_check: string;
  consecutive_fails: number;
  latency: number;
  replication_lag: number;
  potential_data_loss: number;
}

export interface FailoverEvent {
  id: string;
  from_region: string;
  to_region: string;
  reason: string;
  start_time: string;
  end_time?: string;
  duration?: number;
  success: boolean;
  data_loss?: number;
  error_message?: string;
  automatic: boolean;
}

export interface DrillResult {
  id: string;
  start_time: string;
  end_time: string;
  duration: number;
  target_region: string;
  all_passed: boolean;
  checks: DrillCheck[];
}

export interface DrillCheck {
  name: string;
  passed: boolean;
  message: string;
  start_time: string;
  end_time: string;
}

export interface DRAction {
  action: 'failover' | 'failback' | 'drill';
  target?: string;
  reason?: string;
}

