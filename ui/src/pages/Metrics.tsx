import { useQuery } from '@tanstack/react-query';
import { TrendingUp, Activity, Database, AlertCircle, BarChart3 } from 'lucide-react';
import LoadingSpinner from '@/components/ui/LoadingSpinner';
import { appConfig } from '@/core/config';
import { useShards } from '@/features/shard';

export default function Metrics() {
  const { data: shards } = useShards();

  // Fetch metrics from Prometheus endpoints
  const { data: routerMetrics, isLoading: routerMetricsLoading } = useQuery({
    queryKey: ['metrics', 'router'],
    queryFn: async () => {
      const response = await fetch(`${appConfig.getConfig().routerUrl}/metrics`);
      if (!response.ok) throw new Error('Failed to fetch router metrics');
      return response.text();
    },
    refetchInterval: 30000, // Refresh every 30 seconds
    retry: false,
  });

  const { data: managerMetrics, isLoading: managerMetricsLoading } = useQuery({
    queryKey: ['metrics', 'manager'],
    queryFn: async () => {
      const response = await fetch(`${appConfig.getConfig().managerUrl}/metrics`);
      if (!response.ok) throw new Error('Failed to fetch manager metrics');
      return response.text();
    },
    refetchInterval: 30000,
    retry: false,
  });

  // Parse Prometheus metrics (simplified parser)
  const parseMetrics = (metricsText: string) => {
    if (!metricsText) return null;

    const lines = metricsText.split('\n');
    const parsed: Record<string, number> = {};

    for (const line of lines) {
      if (line.startsWith('#') || !line.trim()) continue;

      const match = line.match(/^(\w+)\s+([\d.]+)/);
      if (match) {
        const [, name, value] = match;
        parsed[name] = parseFloat(value);
      }
    }

    return parsed;
  };

  const routerMetricsData = routerMetrics ? parseMetrics(routerMetrics) : null;
  const managerMetricsData = managerMetrics ? parseMetrics(managerMetrics) : null;

  const isLoading = routerMetricsLoading || managerMetricsLoading;
  const hasMetrics = routerMetricsData || managerMetricsData;

  // Calculate stats from real metrics
  const totalQueries = routerMetricsData?.['shard_queries_total'] || 0;
  const avgLatency = routerMetricsData?.['shard_query_duration_seconds'] ?
    (routerMetricsData['shard_query_duration_seconds'] * 1000).toFixed(1) : '0';
  const errorRate = routerMetricsData?.['shard_queries_total'] && routerMetricsData?.['shard_queries_total'] > 0 ?
    ((routerMetricsData['shard_queries_total'] - (routerMetricsData['shard_queries_total'] || 0)) / routerMetricsData['shard_queries_total'] * 100).toFixed(2) : '0';
  const activeShards = shards?.filter(s => s.status === 'active').length || 0;

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <LoadingSpinner size="lg" />
      </div>
    );
  }

  if (!hasMetrics) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white">Metrics</h1>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
            System performance and operational metrics
          </p>
        </div>

        <div className="card">
          <div className="text-center py-12">
            <AlertCircle className="mx-auto h-12 w-12 text-gray-400 dark:text-gray-500" />
            <h3 className="mt-4 text-lg font-medium text-gray-900 dark:text-white">No Metrics Available</h3>
            <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">
              Metrics endpoints are not accessible or no data has been collected yet.
            </p>
            <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
              Ensure metrics endpoints are running:
            </p>
            <ul className="mt-2 text-sm text-gray-500 dark:text-gray-400 space-y-1">
              <li>• Router metrics: {appConfig.getConfig().routerUrl}/metrics</li>
              <li>• Manager metrics: {appConfig.getConfig().managerUrl}/metrics</li>
            </ul>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">Metrics</h1>
        <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
          System performance and operational metrics
        </p>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <div className="card">
          <div className="flex items-center">
            <div className="flex-shrink-0 bg-blue-50 rounded-lg p-3">
              <Activity className="h-6 w-6 text-blue-600" />
            </div>
            <div className="ml-4">
              <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Total Queries</p>
              <p className="text-2xl font-semibold text-gray-900 dark:text-white">
                {totalQueries.toLocaleString()}
              </p>
            </div>
          </div>
        </div>
        <div className="card">
          <div className="flex items-center">
            <div className="flex-shrink-0 bg-green-50 rounded-lg p-3">
              <TrendingUp className="h-6 w-6 text-green-600" />
            </div>
            <div className="ml-4">
              <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Avg Latency</p>
              <p className="text-2xl font-semibold text-gray-900 dark:text-white">{avgLatency}ms</p>
            </div>
          </div>
        </div>
        <div className="card">
          <div className="flex items-center">
            <div className="flex-shrink-0 bg-yellow-50 rounded-lg p-3">
              <BarChart3 className="h-6 w-6 text-yellow-600" />
            </div>
            <div className="ml-4">
              <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Error Rate</p>
              <p className="text-2xl font-semibold text-gray-900 dark:text-white">{errorRate}%</p>
            </div>
          </div>
        </div>
        <div className="card">
          <div className="flex items-center">
            <div className="flex-shrink-0 bg-purple-50 rounded-lg p-3">
              <Database className="h-6 w-6 text-purple-600" />
            </div>
            <div className="ml-4">
              <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Active Shards</p>
              <p className="text-2xl font-semibold text-gray-900 dark:text-white">{activeShards}</p>
            </div>
          </div>
        </div>
      </div>

      {/* Metrics Data Display */}
      <div className="card">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Raw Metrics Data</h2>
        <div className="space-y-4">
          {routerMetricsData && (
            <div>
              <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Router Metrics</h3>
              <pre className="bg-gray-50 dark:bg-gray-800 p-4 rounded-lg text-xs overflow-x-auto text-gray-900 dark:text-gray-300">
                {JSON.stringify(routerMetricsData, null, 2)}
              </pre>
            </div>
          )}
          {managerMetricsData && (
            <div>
              <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Manager Metrics</h3>
              <pre className="bg-gray-50 dark:bg-gray-800 p-4 rounded-lg text-xs overflow-x-auto text-gray-900 dark:text-gray-300">
                {JSON.stringify(managerMetricsData, null, 2)}
              </pre>
            </div>
          )}
        </div>
      </div>

      {(!routerMetricsData && !managerMetricsData) && (
        <div className="card bg-yellow-50 border-yellow-200 dark:bg-yellow-900/20 dark:border-yellow-800">
          <p className="text-sm text-yellow-800 dark:text-yellow-400">
            <strong>Note:</strong> Metrics are being collected. Charts will appear once sufficient data is available.
            Metrics endpoints: Router ({appConfig.getConfig().routerUrl}/metrics),
            Manager ({appConfig.getConfig().managerUrl}/metrics)
          </p>
        </div>
      )}
    </div>
  );
}

