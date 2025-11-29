export interface Branch {
  id: string;
  name: string;
  parent_db_id: string;
  parent_db_name: string;
  status: 'creating' | 'ready' | 'failed' | 'deleting';
  connection_string: string;
  created_at: string;
  updated_at: string;
  metadata?: Record<string, unknown>;
}

export interface CreateBranchRequest {
  name: string;
}

