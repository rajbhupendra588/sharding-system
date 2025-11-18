import { useQuery } from '@tanstack/react-query';
import { CheckCircle, XCircle, AlertTriangle, RefreshCw } from 'lucide-react';
import { apiClient } from '@/lib/api-client';
import LoadingSpinner from '@/components/ui/LoadingSpinner';
import StatusBadge from '@/components/ui/StatusBadge';
import Button from '@/components/ui/Button';
import { formatDuration } from '@/lib/utils';

export default function Health() {
  const { data: managerHealth, isLoading: managerLoading, refetch: refetchManager } = useQuery({
    queryKey: ['manager-health'],
    queryFn: () => apiClient.getManagerHealth(),
    refetchInterval: 5000,
  });

  const { data: routerHealth, isLoading: routerLoading, refetch: refetchRouter } = useQuery({
    queryKey: ['router-health'],
    queryFn: () => apiClient.getRouterHealth(),
    refetchInterval: 5000,
  });

  const { data: shards } = useQuery({
    queryKey: ['shards'],
    queryFn: () => apiClient.listShards(),
    refetchInterval: 10000,
  });

  const isLoading = managerLoading || routerLoading;

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <LoadingSpinner size="lg" />
      </div>
    );
  }

  const allHealthy = managerHealth?.status === 'healthy' && routerHealth?.status === 'healthy';

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-gray-900">Health Status</h1>
        <div className="flex items-center space-x-2">
          <Button variant="secondary" onClick={() => { refetchManager(); refetchRouter(); }}>
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh
          </Button>
        </div>
      </div>

      {/* System Status */}
      <div className={`card ${allHealthy ? 'bg-green-50 border-green-200' : 'bg-yellow-50 border-yellow-200'}`}>
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            {allHealthy ? (
              <CheckCircle className="h-8 w-8 text-green-600" />
            ) : (
              <AlertTriangle className="h-8 w-8 text-yellow-600" />
            )}
            <div>
              <h2 className="text-xl font-semibold text-gray-900">System Status</h2>
              <p className="text-sm text-gray-600">
                {allHealthy ? 'All systems operational' : 'Some systems may be degraded'}
              </p>
            </div>
          </div>
          <StatusBadge status={allHealthy ? 'healthy' : 'degraded'} />
        </div>
      </div>

      {/* Service Health */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Manager Health */}
        <div className="card">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold text-gray-900">Manager (Control Plane)</h2>
            <StatusBadge status={managerHealth?.status || 'unknown'} />
          </div>
          <dl className="space-y-2">
            <div className="flex justify-between">
              <dt className="text-sm text-gray-500">Status</dt>
              <dd className="text-sm font-medium text-gray-900">{managerHealth?.status || 'Unknown'}</dd>
            </div>
            {managerHealth?.version && (
              <div className="flex justify-between">
                <dt className="text-sm text-gray-500">Version</dt>
                <dd className="text-sm text-gray-900">{managerHealth.version}</dd>
              </div>
            )}
            {managerHealth?.uptime_seconds && (
              <div className="flex justify-between">
                <dt className="text-sm text-gray-500">Uptime</dt>
                <dd className="text-sm text-gray-900">
                  {formatDuration(managerHealth.uptime_seconds * 1000)}
                </dd>
              </div>
            )}
          </dl>
        </div>

        {/* Router Health */}
        <div className="card">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold text-gray-900">Router (Data Plane)</h2>
            <StatusBadge status={routerHealth?.status || 'unknown'} />
          </div>
          <dl className="space-y-2">
            <div className="flex justify-between">
              <dt className="text-sm text-gray-500">Status</dt>
              <dd className="text-sm font-medium text-gray-900">{routerHealth?.status || 'Unknown'}</dd>
            </div>
            {routerHealth?.version && (
              <div className="flex justify-between">
                <dt className="text-sm text-gray-500">Version</dt>
                <dd className="text-sm text-gray-900">{routerHealth.version}</dd>
              </div>
            )}
            {routerHealth?.uptime_seconds && (
              <div className="flex justify-between">
                <dt className="text-sm text-gray-500">Uptime</dt>
                <dd className="text-sm text-gray-900">
                  {formatDuration(routerHealth.uptime_seconds * 1000)}
                </dd>
              </div>
            )}
          </dl>
        </div>
      </div>

      {/* Shard Health Summary */}
      {shards && shards.length > 0 && (
        <div className="card">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Shard Health Summary</h2>
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div className="bg-green-50 rounded-lg p-4">
              <div className="flex items-center space-x-2">
                <CheckCircle className="h-5 w-5 text-green-600" />
                <span className="text-sm font-medium text-gray-900">Healthy</span>
              </div>
              <p className="text-2xl font-bold text-green-600 mt-2">
                {shards.filter((s) => s.status === 'active').length}
              </p>
            </div>
            <div className="bg-yellow-50 rounded-lg p-4">
              <div className="flex items-center space-x-2">
                <AlertTriangle className="h-5 w-5 text-yellow-600" />
                <span className="text-sm font-medium text-gray-900">Degraded</span>
              </div>
              <p className="text-2xl font-bold text-yellow-600 mt-2">
                {shards.filter((s) => s.status === 'migrating' || s.status === 'readonly').length}
              </p>
            </div>
            <div className="bg-red-50 rounded-lg p-4">
              <div className="flex items-center space-x-2">
                <XCircle className="h-5 w-5 text-red-600" />
                <span className="text-sm font-medium text-gray-900">Unhealthy</span>
              </div>
              <p className="text-2xl font-bold text-red-600 mt-2">
                {shards.filter((s) => s.status === 'inactive').length}
              </p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

