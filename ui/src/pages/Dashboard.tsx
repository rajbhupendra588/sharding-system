import { Link } from 'react-router-dom';
import { Database, Activity, AlertCircle, CheckCircle, Clock } from 'lucide-react';
import { useShards } from '@/features/shard';
import { useSystemHealth } from '@/features/health';
import LoadingSpinner from '@/components/ui/LoadingSpinner';
import StatusBadge from '@/components/ui/StatusBadge';
import { formatRelativeTime } from '@/shared/utils';

export default function Dashboard() {
  const { data: shards, isLoading: shardsLoading } = useShards();
  const { isHealthy } = useSystemHealth();

  if (shardsLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <LoadingSpinner size="lg" />
      </div>
    );
  }

  const activeShards = shards?.filter((s) => s.status === 'active') || [];
  const unhealthyShards = shards?.filter((s) => s.status !== 'active') || [];
  const totalShards = shards?.length || 0;

  const stats = [
    {
      name: 'Total Shards',
      value: totalShards,
      icon: Database,
      color: 'text-blue-600',
      bgColor: 'bg-blue-50',
      link: '/shards',
    },
    {
      name: 'Active Shards',
      value: activeShards.length,
      icon: CheckCircle,
      color: 'text-green-600',
      bgColor: 'bg-green-50',
      link: '/shards',
    },
    {
      name: 'Unhealthy Shards',
      value: unhealthyShards.length,
      icon: AlertCircle,
      color: 'text-red-600',
      bgColor: 'bg-red-50',
      link: '/health',
    },
    {
      name: 'System Status',
      value: isHealthy ? 'Healthy' : 'Degraded',
      icon: Activity,
      color: isHealthy ? 'text-green-600' : 'text-yellow-600',
      bgColor: isHealthy ? 'bg-green-50' : 'bg-yellow-50',
      link: '/health',
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">Dashboard</h1>
        <div className="flex items-center space-x-4">
          <StatusBadge
            status={isHealthy ? 'healthy' : 'degraded'}
          />
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
        {stats.map((stat) => (
          <Link
            key={stat.name}
            to={stat.link}
            className="card hover:shadow-md transition-shadow"
          >
            <div className="flex items-center">
              <div className={`flex-shrink-0 ${stat.bgColor} rounded-lg p-3`}>
                <stat.icon className={`h-6 w-6 ${stat.color}`} />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-600 dark:text-gray-400">{stat.name}</p>
                <p className={`text-2xl font-semibold ${stat.color}`}>
                  {stat.value}
                </p>
              </div>
            </div>
          </Link>
        ))}
      </div>

      {/* Recent Shards */}
      <div className="card">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-semibold text-gray-900 dark:text-white">Recent Shards</h2>
          <Link
            to="/shards"
            className="text-sm font-medium text-primary-600 hover:text-primary-700"
          >
            View all →
          </Link>
        </div>
        {shards && shards.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
              <thead className="bg-gray-50 dark:bg-gray-900/50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Name
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Replicas
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    Updated
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                {shards.slice(0, 5).map((shard) => (
                  <tr
                    key={shard.id}
                    className="hover:bg-gray-50 dark:hover:bg-gray-700/50 cursor-pointer"
                    onClick={() => window.location.href = `/shards/${shard.id}`}
                  >
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm font-medium text-gray-900 dark:text-white">{shard.name}</div>
                      <div className="text-sm text-gray-500 dark:text-gray-400">{shard.id}</div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <StatusBadge status={shard.status} />
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-300">
                      {shard.replicas?.length || 0}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                      {formatRelativeTime(shard.updated_at)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="text-center py-12">
            <Database className="mx-auto h-12 w-12 text-gray-400" />
            <h3 className="mt-2 text-sm font-medium text-gray-900 dark:text-white">No shards</h3>
            <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">Get started by creating a new shard.</p>
            <div className="mt-6">
              <Link
                to="/shards"
                className="btn btn-primary"
              >
                Create Shard
              </Link>
            </div>
          </div>
        )}
      </div>

      {/* Quick Actions */}
      <div className="grid grid-cols-1 gap-6 sm:grid-cols-3">
        <Link
          to="/query"
          className="card hover:shadow-md transition-shadow"
        >
          <div className="flex items-center">
            <div className="flex-shrink-0 bg-primary-50 rounded-lg p-3">
              <Clock className="h-6 w-6 text-primary-600" />
            </div>
            <div className="ml-4">
              <h3 className="text-lg font-medium text-gray-900 dark:text-white">Execute Query</h3>
              <p className="text-sm text-gray-500 dark:text-gray-400">Run SQL queries against shards</p>
            </div>
          </div>
        </Link>
        <Link
          to="/resharding"
          className="card hover:shadow-md transition-shadow"
        >
          <div className="flex items-center">
            <div className="flex-shrink-0 bg-purple-50 rounded-lg p-3">
              <Activity className="h-6 w-6 text-purple-600" />
            </div>
            <div className="ml-4">
              <h3 className="text-lg font-medium text-gray-900 dark:text-white">Resharding</h3>
              <p className="text-sm text-gray-500 dark:text-gray-400">Split or merge shards</p>
            </div>
          </div>
        </Link>
        <Link
          to="/metrics"
          className="card hover:shadow-md transition-shadow"
        >
          <div className="flex items-center">
            <div className="flex-shrink-0 bg-indigo-50 rounded-lg p-3">
              <Activity className="h-6 w-6 text-indigo-600" />
            </div>
            <div className="ml-4">
              <h3 className="text-lg font-medium text-gray-900 dark:text-white">Metrics</h3>
              <p className="text-sm text-gray-500 dark:text-gray-400">View system metrics</p>
            </div>
          </div>
        </Link>
      </div>
    </div>
  );
}

