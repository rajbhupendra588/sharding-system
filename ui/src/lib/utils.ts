/**
 * Legacy Utils - Deprecated
 * Use @/shared/utils and @/shared/lib instead
 * @deprecated
 */

// Re-export from new modular structure for backward compatibility
export { cn } from '@/shared/lib';
export {
  formatBytes,
  formatDuration,
  formatDate,
  formatRelativeTime,
} from '@/shared/utils/formatting';
export {
  getStatusColor,
  getStatusBadgeColor,
} from '@/shared/utils/status';
export { debounce, throttle } from '@/shared/utils/debounce';

