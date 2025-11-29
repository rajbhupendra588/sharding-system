export interface PostgresStats {
  database_id: string;
  database_name: string;
  size_bytes: number;
  collected_at: string;
  connections: ConnectionStats;
  queries: QueryStats;
  replication: ReplicationStats;
  tables: TableStats;
  indexes: IndexStats;
  locks: LockStats;
  bg_writer: BGWriterStats;
}

export interface ConnectionStats {
  total: number;
  active: number;
  idle: number;
  idle_in_transaction: number;
  waiting: number;
  max_connections: number;
  percent_used: number;
  by_state: Record<string, number>;
  by_application: Record<string, number>;
}

export interface QueryStats {
  total_queries: number;
  queries_per_second: number;
  avg_query_time_ms: number;
  max_query_time_ms: number;
  slow_queries: number;
  cache_hit_ratio: number;
  top_queries?: TopQuery[];
}

export interface TopQuery {
  query: string;
  calls: number;
  total_time_ms: number;
  mean_time_ms: number;
  rows: number;
  shared_blocks_hit: number;
}

export interface ReplicationStats {
  is_replica: boolean;
  replication_lag_seconds: number;
  replica_count: number;
  wal_position: string;
  replay_lag_bytes: number;
  replicas?: ReplicaInfo[];
}

export interface ReplicaInfo {
  client_addr: string;
  state: string;
  sent_lag_bytes: number;
  write_lag_bytes: number;
  flush_lag_bytes: number;
  replay_lag_bytes: number;
  sync_state: string;
}

export interface TableStats {
  total_tables: number;
  total_rows: number;
  live_tuples: number;
  dead_tuples: number;
  sequential_scans: number;
  index_scans: number;
  seq_scan_ratio: number;
  largest_tables?: TableInfo[];
}

export interface TableInfo {
  schema: string;
  table_name: string;
  rows: number;
  size_bytes: number;
  seq_scans: number;
  idx_scans: number;
}

export interface IndexStats {
  total_indexes: number;
  index_size_bytes: number;
  index_hit_ratio: number;
  unused_indexes: number;
  duplicate_indexes: number;
}

export interface LockStats {
  total: number;
  granted: number;
  waiting: number;
  deadlocks: number;
  by_type: Record<string, number>;
  by_mode: Record<string, number>;
}

export interface BGWriterStats {
  checkpoints_timed: number;
  checkpoints_requested: number;
  buffers_checkpoint: number;
  buffers_clean: number;
  maxwritten_clean: number;
  buffers_backend: number;
  buffers_backend_fsync: number;
  buffers_alloc: number;
}

