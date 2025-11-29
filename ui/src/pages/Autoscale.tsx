import { useState } from 'react';
import { Activity, TrendingUp, TrendingDown, Settings, Power, PowerOff } from 'lucide-react';
import { useAutoscaleStatus, useHotShards, useColdShards, useThresholds, useEnableAutoscale, useDisableAutoscale, useAllMetrics } from '@/features/autoscale';
import LoadingSpinner from '@/components/ui/LoadingSpinner';
import Button from '@/components/ui/Button';
import StatusBadge from '@/components/ui/StatusBadge';
// Card component - using div with card class
const Card = ({ children, className = '' }: { children: React.ReactNode; className?: string }) => (
  <div className={`card ${className}`}>{children}</div>
);
import { Table, TableHead, TableHeader, TableBody, TableRow, TableCell } from '@/components/ui/Table';
import { Link } from 'react-router-dom';

export default function Autoscale() {
  const [showThresholds, setShowThresholds] = useState(false);

  const { data: status, isLoading: statusLoading } = useAutoscaleStatus();
  const { data: hotShards, isLoading: hotLoading } = useHotShards();
  const { data: coldShards, isLoading: coldLoading } = useColdShards();
  const { data: thresholds } = useThresholds();
  const { data: allMetrics } = useAllMetrics();

  const enableMutation = useEnableAutoscale();
  const disableMutation = useDisableAutoscale();

  const isLoading = statusLoading || hotLoading || coldLoading;

  return (
    <div className="space-y-6 p-4 sm:p-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-white">Auto-Scaling</h1>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
            Automatic shard splitting based on load metrics
          </p>
        </div>
        <div className="flex items-center gap-2">
          {status?.enabled ? (
            <Button
              variant="danger"
              onClick={() => disableMutation.mutate()}
              disabled={disableMutation.isPending}
            >
              <PowerOff className="h-4 w-4 mr-2" />
              Disable
            </Button>
          ) : (
            <Button
              onClick={() => enableMutation.mutate()}
              disabled={enableMutation.isPending}
            >
              <Power className="h-4 w-4 mr-2" />
              Enable
            </Button>
          )}
          <Button variant="outline" onClick={() => setShowThresholds(!showThresholds)}>
            <Settings className="h-4 w-4 mr-2" />
            Thresholds
          </Button>
        </div>
      </div>

      {/* Status Card */}
      <Card>
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Status</h3>
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
              {status?.enabled ? 'Auto-scaling is active' : 'Auto-scaling is disabled'}
            </p>
          </div>
          <StatusBadge status={status?.enabled ? 'success' : 'default'} />
        </div>
      </Card>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Card>
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500 dark:text-gray-400">Hot Shards</p>
              <p className="text-2xl font-bold text-gray-900 dark:text-white">
                {hotShards?.length || 0}
              </p>
            </div>
            <TrendingUp className="h-8 w-8 text-red-500" />
          </div>
        </Card>

        <Card>
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500 dark:text-gray-400">Cold Shards</p>
              <p className="text-2xl font-bold text-gray-900 dark:text-white">
                {coldShards?.length || 0}
              </p>
            </div>
            <TrendingDown className="h-8 w-8 text-blue-500" />
          </div>
        </Card>

        <Card>
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500 dark:text-gray-400">Total Shards</p>
              <p className="text-2xl font-bold text-gray-900 dark:text-white">
                {allMetrics ? Object.keys(allMetrics).length : 0}
              </p>
            </div>
            <Activity className="h-8 w-8 text-gray-500" />
          </div>
        </Card>
      </div>

      {/* Thresholds */}
      {showThresholds && thresholds && (
        <Card>
          <h3 className="text-lg font-semibold mb-4 text-gray-900 dark:text-white">Detection Thresholds</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="text-sm font-medium text-gray-500 dark:text-gray-400">Max Query Rate</label>
              <p className="text-gray-900 dark:text-white">{thresholds.max_query_rate.toLocaleString()} qps</p>
            </div>
            <div>
              <label className="text-sm font-medium text-gray-500 dark:text-gray-400">Max CPU Usage</label>
              <p className="text-gray-900 dark:text-white">{thresholds.max_cpu_usage}%</p>
            </div>
            <div>
              <label className="text-sm font-medium text-gray-500 dark:text-gray-400">Max Memory Usage</label>
              <p className="text-gray-900 dark:text-white">{thresholds.max_memory_usage}%</p>
            </div>
            <div>
              <label className="text-sm font-medium text-gray-500 dark:text-gray-400">Max Storage Usage</label>
              <p className="text-gray-900 dark:text-white">{thresholds.max_storage_usage}%</p>
            </div>
            <div>
              <label className="text-sm font-medium text-gray-500 dark:text-gray-400">Max Connections</label>
              <p className="text-gray-900 dark:text-white">{thresholds.max_connections.toLocaleString()}</p>
            </div>
            <div>
              <label className="text-sm font-medium text-gray-500 dark:text-gray-400">Max Latency</label>
              <p className="text-gray-900 dark:text-white">{thresholds.max_latency_ms}ms</p>
            </div>
          </div>
        </Card>
      )}

      {/* Hot Shards */}
      {hotShards && hotShards.length > 0 && (
        <Card>
          <h3 className="text-lg font-semibold mb-4 text-gray-900 dark:text-white">Hot Shards (Auto-Split Candidates)</h3>
          <div className="space-y-2">
            {hotShards.map((shardID) => {
              const metrics = allMetrics?.[shardID];
              return (
                <div
                  key={shardID}
                  className="flex items-center justify-between p-3 bg-red-50 dark:bg-red-900/20 rounded-lg"
                >
                  <div>
                    <Link to={`/shards/${shardID}`} className="font-medium text-gray-900 dark:text-white hover:underline">
                      {shardID}
                    </Link>
                    {metrics && (
                      <div className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                        {(metrics.query_rate || 0).toFixed(0)} qps • {(metrics.cpu_usage || 0).toFixed(1)}% CPU • {(metrics.storage_usage || 0).toFixed(1)}% Storage
                      </div>
                    )}
                  </div>
                  <StatusBadge status="error" />
                </div>
              );
            })}
          </div>
        </Card>
      )}

      {/* Cold Shards */}
      {coldShards && coldShards.length > 0 && (
        <Card>
          <h3 className="text-lg font-semibold mb-4 text-gray-900 dark:text-white">Cold Shards (Merge Candidates)</h3>
          <div className="space-y-2">
            {coldShards.map((shardID) => {
              const metrics = allMetrics?.[shardID];
              return (
                <div
                  key={shardID}
                  className="flex items-center justify-between p-3 bg-blue-50 dark:bg-blue-900/20 rounded-lg"
                >
                  <div>
                    <Link to={`/shards/${shardID}`} className="font-medium text-gray-900 dark:text-white hover:underline">
                      {shardID}
                    </Link>
                    {metrics && (
                      <div className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                        {(metrics.query_rate || 0).toFixed(0)} qps • {(metrics.cpu_usage || 0).toFixed(1)}% CPU • {(metrics.storage_usage || 0).toFixed(1)}% Storage
                      </div>
                    )}
                  </div>
                  <StatusBadge status="info" />
                </div>
              );
            })}
          </div>
        </Card>
      )}

      {/* All Metrics Table */}
      {allMetrics && Object.keys(allMetrics).length > 0 ? (
        <Card>
          <h3 className="text-lg font-semibold mb-4 text-gray-900 dark:text-white">All Shard Metrics</h3>
          <Table>
            <TableHead>
              <TableHeader>Shard ID</TableHeader>
              <TableHeader>Query Rate</TableHeader>
              <TableHeader>CPU</TableHeader>
              <TableHeader>Memory</TableHeader>
              <TableHeader>Storage</TableHeader>
              <TableHeader>Connections</TableHeader>
              <TableHeader>Latency</TableHeader>
            </TableHead>
            <TableBody>
              {Object.entries(allMetrics).map(([shardID, metrics]) => {
                // Handle metrics from API (snake_case)
                const queryRate = metrics?.query_rate ?? 0;
                const cpuUsage = metrics?.cpu_usage ?? 0;
                const memoryUsage = metrics?.memory_usage ?? 0;
                const storageUsage = metrics?.storage_usage ?? 0;
                const connectionCount = metrics?.connection_count ?? 0;
                const avgLatencyMs = metrics?.avg_latency_ms ?? 0;

                return (
                  <TableRow key={shardID}>
                    <TableCell>
                      <Link to={`/shards/${shardID}`} className="text-blue-600 hover:underline">
                        {shardID}
                      </Link>
                    </TableCell>
                    <TableCell>{queryRate.toFixed(0)} qps</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <div className="w-16 bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                          <div
                            className={`h-2 rounded-full ${cpuUsage > 80 ? 'bg-red-500' : cpuUsage > 60 ? 'bg-yellow-500' : 'bg-green-500'
                              }`}
                            style={{ width: `${Math.min(cpuUsage, 100)}%` }}
                          />
                        </div>
                        <span>{cpuUsage.toFixed(1)}%</span>
                      </div>
                    </TableCell>
                    <TableCell>{memoryUsage.toFixed(1)}%</TableCell>
                    <TableCell>{storageUsage.toFixed(1)}%</TableCell>
                    <TableCell>{connectionCount}</TableCell>
                    <TableCell>{avgLatencyMs.toFixed(1)}ms</TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </Card>
      ) : (
        <Card>
          <div className="text-center py-12">
            <Activity className="h-12 w-12 mx-auto text-gray-400 mb-4" />
            <p className="text-gray-600 dark:text-gray-400">
              {isLoading ? 'Loading metrics...' : 'No metrics available yet. Metrics are collected every 10 seconds.'}
            </p>
          </div>
        </Card>
      )}

      {isLoading && (
        <div className="flex justify-center py-12">
          <LoadingSpinner />
        </div>
      )}
    </div>
  );
}

