/**
 * Resharding Feature Types
 */

import { CreateShardRequest } from '@/features/shard/types';

export interface ReshardJob {
  id: string;
  type: ReshardType;
  source_shards: string[];
  target_shards: string[];
  status: ReshardStatus;
  progress: number; // 0.0 to 1.0
  started_at: string;
  completed_at?: string;
  error_message?: string;
  keys_migrated: number;
  total_keys: number;
}

export type ReshardType = 'split' | 'merge';
export type ReshardStatus =
  | 'pending'
  | 'precopy'
  | 'deltasync'
  | 'cutover'
  | 'validation'
  | 'completed'
  | 'failed';

export interface SplitRequest {
  source_shard_id: string;
  target_shards: CreateShardRequest[];
  split_point?: number;
}

export interface MergeRequest {
  source_shard_ids: string[];
  target_shard: CreateShardRequest;
}

