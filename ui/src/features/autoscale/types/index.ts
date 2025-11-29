export interface Thresholds {
  max_query_rate: number;
  max_cpu_usage: number;
  max_memory_usage: number;
  max_storage_usage: number;
  max_connections: number;
  max_latency_ms: number;
  min_query_rate: number;
  min_cpu_usage: number;
  min_storage_usage: number;
}

export interface ShardMetrics {
  shard_id: string;
  query_rate: number;
  connection_count: number;
  cpu_usage: number;
  memory_usage: number;
  storage_usage: number;
  avg_latency_ms: number;
  error_rate: number;
  timestamp: string;
}

export interface AutoscaleStatus {
  enabled: boolean;
}

