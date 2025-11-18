/**
 * Status Utilities
 * Status badge colors and helpers
 */

export function getStatusColor(status: string): string {
  const statusColors: Record<string, string> = {
    healthy: 'text-green-600 bg-green-50',
    active: 'text-green-600 bg-green-50',
    degraded: 'text-yellow-600 bg-yellow-50',
    unhealthy: 'text-red-600 bg-red-50',
    failed: 'text-red-600 bg-red-50',
    pending: 'text-gray-600 bg-gray-50',
    migrating: 'text-blue-600 bg-blue-50',
    readonly: 'text-orange-600 bg-orange-50',
    inactive: 'text-gray-600 bg-gray-50',
    completed: 'text-green-600 bg-green-50',
    precopy: 'text-blue-600 bg-blue-50',
    deltasync: 'text-blue-600 bg-blue-50',
    cutover: 'text-purple-600 bg-purple-50',
    validation: 'text-indigo-600 bg-indigo-50',
  };
  return statusColors[status.toLowerCase()] || 'text-gray-600 bg-gray-50';
}

export function getStatusBadgeColor(status: string): string {
  const statusColors: Record<string, string> = {
    healthy: 'bg-green-100 text-green-800 border-green-200',
    active: 'bg-green-100 text-green-800 border-green-200',
    degraded: 'bg-yellow-100 text-yellow-800 border-yellow-200',
    unhealthy: 'bg-red-100 text-red-800 border-red-200',
    failed: 'bg-red-100 text-red-800 border-red-200',
    pending: 'bg-gray-100 text-gray-800 border-gray-200',
    migrating: 'bg-blue-100 text-blue-800 border-blue-200',
    readonly: 'bg-orange-100 text-orange-800 border-orange-200',
    inactive: 'bg-gray-100 text-gray-800 border-gray-200',
    completed: 'bg-green-100 text-green-800 border-green-200',
    precopy: 'bg-blue-100 text-blue-800 border-blue-200',
    deltasync: 'bg-blue-100 text-blue-800 border-blue-200',
    cutover: 'bg-purple-100 text-purple-800 border-purple-200',
    validation: 'bg-indigo-100 text-indigo-800 border-indigo-200',
  };
  return statusColors[status.toLowerCase()] || 'bg-gray-100 text-gray-800 border-gray-200';
}

