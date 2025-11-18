/**
 * Shared Types
 * Common types used across the application
 */

export interface ApiError {
  message: string;
  code?: string;
  details?: unknown;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
}

export interface BaseEntity {
  id: string;
  created_at: string;
  updated_at: string;
}

