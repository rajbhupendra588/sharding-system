export interface DatabaseTemplate {
  name: string;
  display_name: string;
  description: string;
  shard_count: number;
  replicas_per_shard: number;
  vnode_count: number;
  auto_scale: boolean;
  estimated_cost: string;
}

export interface Database {
  id: string;
  name: string;
  display_name?: string;
  description?: string;
  template: string;
  shard_key: string;
  client_app_id: string;
  shard_ids: string[];
  status: 'creating' | 'ready' | 'failed' | 'scaling';
  connection_string: string;
  created_at: string;
  updated_at: string;
  metadata?: Record<string, unknown>;
}

export interface CreateDatabaseRequest {
  name: string;
  template?: string;
  shard_key?: string;
  display_name?: string;
  description?: string;
}

export interface DatabaseStatus {
  id: string;
  name: string;
  status: string;
  shard_count: number;
  connection_string: string;
  created_at: string;
  updated_at: string;
}

export interface Backup {
  id: string;
  database_id: string;
  type: 'full' | 'incremental';
  status: 'pending' | 'in_progress' | 'completed' | 'failed';
  size: number;
  storage_path: string;
  created_at: string;
  completed_at?: string;
  error?: string;
}

export interface FailoverStatus {
  enabled: boolean;
  status: string;
}

export interface FailoverEvent {
  id: string;
  shard_id: string;
  old_primary: string;
  new_primary: string;
  reason: string;
  status: 'success' | 'failed' | 'rolled_back' | 'in_progress';
  started_at: string;
  completed_at?: string;
  error?: string;
}

