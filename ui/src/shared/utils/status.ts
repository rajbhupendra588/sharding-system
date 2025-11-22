/**
 * Status Utilities
 * Status badge colors and helpers
 */

export function getStatusColor(status: string): string {
  const statusColors: Record<string, string> = {
    healthy: 'text-green-600 bg-green-50 dark:text-green-400 dark:bg-green-900/20',
    active: 'text-green-600 bg-green-50 dark:text-green-400 dark:bg-green-900/20',
    degraded: 'text-yellow-600 bg-yellow-50 dark:text-yellow-400 dark:bg-yellow-900/20',
    unhealthy: 'text-red-600 bg-red-50 dark:text-red-400 dark:bg-red-900/20',
    failed: 'text-red-600 bg-red-50 dark:text-red-400 dark:bg-red-900/20',
    pending: 'text-gray-600 bg-gray-50 dark:text-gray-400 dark:bg-gray-800',
    migrating: 'text-blue-600 bg-blue-50 dark:text-blue-400 dark:bg-blue-900/20',
    readonly: 'text-orange-600 bg-orange-50 dark:text-orange-400 dark:bg-orange-900/20',
    inactive: 'text-gray-600 bg-gray-50 dark:text-gray-400 dark:bg-gray-800',
    completed: 'text-green-600 bg-green-50 dark:text-green-400 dark:bg-green-900/20',
    precopy: 'text-blue-600 bg-blue-50 dark:text-blue-400 dark:bg-blue-900/20',
    deltasync: 'text-blue-600 bg-blue-50 dark:text-blue-400 dark:bg-blue-900/20',
    cutover: 'text-purple-600 bg-purple-50 dark:text-purple-400 dark:bg-purple-900/20',
    validation: 'text-indigo-600 bg-indigo-50 dark:text-indigo-400 dark:bg-indigo-900/20',
  };
  return statusColors[status.toLowerCase()] || 'text-gray-600 bg-gray-50 dark:text-gray-400 dark:bg-gray-800';
}

export function getStatusBadgeColor(status: string): string {
  const statusColors: Record<string, string> = {
    healthy: 'bg-green-100 text-green-800 border-green-200 dark:bg-green-900/30 dark:text-green-300 dark:border-green-800',
    active: 'bg-green-100 text-green-800 border-green-200 dark:bg-green-900/30 dark:text-green-300 dark:border-green-800',
    degraded: 'bg-yellow-100 text-yellow-800 border-yellow-200 dark:bg-yellow-900/30 dark:text-yellow-300 dark:border-yellow-800',
    unhealthy: 'bg-red-100 text-red-800 border-red-200 dark:bg-red-900/30 dark:text-red-300 dark:border-red-800',
    failed: 'bg-red-100 text-red-800 border-red-200 dark:bg-red-900/30 dark:text-red-300 dark:border-red-800',
    pending: 'bg-gray-100 text-gray-800 border-gray-200 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-700',
    migrating: 'bg-blue-100 text-blue-800 border-blue-200 dark:bg-blue-900/30 dark:text-blue-300 dark:border-blue-800',
    readonly: 'bg-orange-100 text-orange-800 border-orange-200 dark:bg-orange-900/30 dark:text-orange-300 dark:border-orange-800',
    inactive: 'bg-gray-100 text-gray-800 border-gray-200 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-700',
    completed: 'bg-green-100 text-green-800 border-green-200 dark:bg-green-900/30 dark:text-green-300 dark:border-green-800',
    precopy: 'bg-blue-100 text-blue-800 border-blue-200 dark:bg-blue-900/30 dark:text-blue-300 dark:border-blue-800',
    deltasync: 'bg-blue-100 text-blue-800 border-blue-200 dark:bg-blue-900/30 dark:text-blue-300 dark:border-blue-800',
    cutover: 'bg-purple-100 text-purple-800 border-purple-200 dark:bg-purple-900/30 dark:text-purple-300 dark:border-purple-800',
    validation: 'bg-indigo-100 text-indigo-800 border-indigo-200 dark:bg-indigo-900/30 dark:text-indigo-300 dark:border-indigo-800',
  };
  return statusColors[status.toLowerCase()] || 'bg-gray-100 text-gray-800 border-gray-200 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-700';
}

