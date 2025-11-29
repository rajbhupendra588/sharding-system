import { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import {
  Database,
  Activity,
  Clock,
  HardDrive,
  Users,
  Zap,
  RefreshCw,
  BarChart3,
  Lock,
  Table,
  TrendingUp,
  AlertTriangle,
  CheckCircle,
  Gauge,
  ChevronDown,
} from 'lucide-react';
import { usePostgresStats } from '@/features/postgres-stats';
import type { ConnectionStats, QueryStats, TableStats, LockStats } from '@/features/postgres-stats';

const formatBytes = (bytes: number) => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

const formatNumber = (num: number) => {
  if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M';
  if (num >= 1000) return (num / 1000).toFixed(1) + 'K';
  return num.toString();
};

function StatCard({ title, value, subtitle, icon: Icon, color, trend }: {
  title: string;
  value: string | number;
  subtitle?: string;
  icon: any;
  color: string;
  trend?: 'up' | 'down' | 'neutral';
}) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="bg-white dark:bg-gray-800 rounded-xl p-6 border border-gray-200 dark:border-gray-700 shadow-sm"
    >
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm text-gray-500 dark:text-gray-400">{title}</p>
          <p className="text-2xl font-bold text-gray-900 dark:text-white mt-1">{value}</p>
          {subtitle && (
            <p className="text-xs text-gray-400 dark:text-gray-500 mt-1">{subtitle}</p>
          )}
        </div>
        <div className={`p-3 rounded-lg ${color}`}>
          <Icon className="w-6 h-6" />
        </div>
      </div>
      {trend && (
        <div className={`mt-3 flex items-center text-sm ${trend === 'up' ? 'text-green-600' : trend === 'down' ? 'text-red-600' : 'text-gray-500'
          }`}>
          <TrendingUp className={`w-4 h-4 mr-1 ${trend === 'down' ? 'rotate-180' : ''}`} />
          <span>{trend === 'up' ? 'Improving' : trend === 'down' ? 'Needs attention' : 'Stable'}</span>
        </div>
      )}
    </motion.div>
  );
}

function ConnectionsChart({ connections }: { connections: ConnectionStats }) {
  const total = connections?.max_connections ?? 100;
  const used = connections?.total ?? 0;
  const percentage = total > 0 ? (used / total) * 100 : 0;

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: 0.1 }}
      className="bg-white dark:bg-gray-800 rounded-xl p-6 border border-gray-200 dark:border-gray-700 shadow-sm"
    >
      <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
        <Users className="w-5 h-5 text-gray-400" />
        Connection Pool
      </h3>

      <div className="space-y-4">
        {/* Main gauge */}
        <div className="relative h-32 flex items-center justify-center">
          <svg className="w-32 h-32 transform -rotate-90">
            <circle
              cx="64"
              cy="64"
              r="56"
              fill="none"
              stroke="currentColor"
              strokeWidth="12"
              className="text-gray-200 dark:text-gray-700"
            />
            <circle
              cx="64"
              cy="64"
              r="56"
              fill="none"
              stroke="currentColor"
              strokeWidth="12"
              strokeDasharray={`${percentage * 3.52} 352`}
              className={percentage > 80 ? 'text-red-500' : percentage > 60 ? 'text-yellow-500' : 'text-green-500'}
            />
          </svg>
          <div className="absolute text-center">
            <p className="text-2xl font-bold text-gray-900 dark:text-white">{used}</p>
            <p className="text-xs text-gray-500">of {total}</p>
          </div>
        </div>

        {/* Connection breakdown */}
        <div className="grid grid-cols-2 gap-3">
          <div className="bg-green-50 dark:bg-green-900/20 rounded-lg p-3">
            <p className="text-xs text-green-600 dark:text-green-400">Active</p>
            <p className="text-lg font-semibold text-green-700 dark:text-green-300">{connections?.active ?? 0}</p>
          </div>
          <div className="bg-blue-50 dark:bg-blue-900/20 rounded-lg p-3">
            <p className="text-xs text-blue-600 dark:text-blue-400">Idle</p>
            <p className="text-lg font-semibold text-blue-700 dark:text-blue-300">{connections?.idle ?? 0}</p>
          </div>
          <div className="bg-yellow-50 dark:bg-yellow-900/20 rounded-lg p-3">
            <p className="text-xs text-yellow-600 dark:text-yellow-400">Idle in TX</p>
            <p className="text-lg font-semibold text-yellow-700 dark:text-yellow-300">{connections?.idle_in_transaction ?? 0}</p>
          </div>
          <div className="bg-red-50 dark:bg-red-900/20 rounded-lg p-3">
            <p className="text-xs text-red-600 dark:text-red-400">Waiting</p>
            <p className="text-lg font-semibold text-red-700 dark:text-red-300">{connections?.waiting ?? 0}</p>
          </div>
        </div>
      </div>
    </motion.div>
  );
}

function QueryPerformance({ queries }: { queries: QueryStats }) {
  const cacheHitRatio = queries?.cache_hit_ratio ?? 0;
  const queriesPerSecond = queries?.queries_per_second ?? 0;
  const avgQueryTime = queries?.avg_query_time_ms ?? 0;
  const slowQueries = queries?.slow_queries ?? 0;
  const totalQueries = queries?.total_queries ?? 0;

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: 0.2 }}
      className="bg-white dark:bg-gray-800 rounded-xl p-6 border border-gray-200 dark:border-gray-700 shadow-sm"
    >
      <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
        <Zap className="w-5 h-5 text-gray-400" />
        Query Performance
      </h3>

      <div className="space-y-4">
        {/* Cache hit ratio gauge */}
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-sm text-gray-500 dark:text-gray-400">Cache Hit Ratio</span>
            <span className={`text-sm font-medium ${cacheHitRatio >= 99 ? 'text-green-600' : cacheHitRatio >= 95 ? 'text-yellow-600' : 'text-red-600'}`}>
              {cacheHitRatio.toFixed(2)}%
            </span>
          </div>
          <div className="h-2 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
            <div
              className={`h-full rounded-full transition-all ${cacheHitRatio >= 99 ? 'bg-green-500' : cacheHitRatio >= 95 ? 'bg-yellow-500' : 'bg-red-500'}`}
              style={{ width: `${Math.min(cacheHitRatio, 100)}%` }}
            />
          </div>
        </div>

        {/* Query stats */}
        <div className="grid grid-cols-2 gap-4 pt-4 border-t border-gray-200 dark:border-gray-700">
          <div>
            <p className="text-xs text-gray-500 dark:text-gray-400">Queries/sec</p>
            <p className="text-xl font-semibold text-gray-900 dark:text-white">{queriesPerSecond.toFixed(1)}</p>
          </div>
          <div>
            <p className="text-xs text-gray-500 dark:text-gray-400">Avg Time</p>
            <p className="text-xl font-semibold text-gray-900 dark:text-white">{avgQueryTime.toFixed(2)}ms</p>
          </div>
          <div>
            <p className="text-xs text-gray-500 dark:text-gray-400">Slow Queries</p>
            <p className={`text-xl font-semibold ${slowQueries > 0 ? 'text-red-600' : 'text-gray-900 dark:text-white'}`}>
              {formatNumber(slowQueries)}
            </p>
          </div>
          <div>
            <p className="text-xs text-gray-500 dark:text-gray-400">Total Queries</p>
            <p className="text-xl font-semibold text-gray-900 dark:text-white">{formatNumber(totalQueries)}</p>
          </div>
        </div>
      </div>
    </motion.div>
  );
}

function TableOverview({ tables }: { tables: TableStats }) {
  const deadTuples = tables?.dead_tuples ?? 0;
  const liveTuples = tables?.live_tuples ?? 1;
  const bloatRatio = deadTuples / Math.max(liveTuples, 1) * 100;

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: 0.3 }}
      className="bg-white dark:bg-gray-800 rounded-xl p-6 border border-gray-200 dark:border-gray-700 shadow-sm"
    >
      <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
        <Table className="w-5 h-5 text-gray-400" />
        Table Statistics
      </h3>

      <div className="space-y-4">
        <div className="grid grid-cols-3 gap-4">
          <div className="text-center">
            <p className="text-2xl font-bold text-gray-900 dark:text-white">{tables?.total_tables ?? 0}</p>
            <p className="text-xs text-gray-500 dark:text-gray-400">Tables</p>
          </div>
          <div className="text-center">
            <p className="text-2xl font-bold text-gray-900 dark:text-white">{formatNumber(tables?.total_rows ?? 0)}</p>
            <p className="text-xs text-gray-500 dark:text-gray-400">Total Rows</p>
          </div>
          <div className="text-center">
            <p className={`text-2xl font-bold ${bloatRatio > 20 ? 'text-red-600' : 'text-gray-900 dark:text-white'}`}>
              {bloatRatio.toFixed(1)}%
            </p>
            <p className="text-xs text-gray-500 dark:text-gray-400">Dead Tuples</p>
          </div>
        </div>

        {/* Scan ratio */}
        <div className="pt-4 border-t border-gray-200 dark:border-gray-700 space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-sm text-gray-500 dark:text-gray-400">Index vs Sequential Scans</span>
          </div>
          <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden flex">
            <div
              className="h-full bg-green-500"
              style={{ width: `${100 - (tables?.seq_scan_ratio ?? 0)}%` }}
              title={`Index scans: ${formatNumber(tables?.index_scans ?? 0)}`}
            />
            <div
              className="h-full bg-yellow-500"
              style={{ width: `${tables?.seq_scan_ratio ?? 0}%` }}
              title={`Sequential scans: ${formatNumber(tables?.sequential_scans ?? 0)}`}
            />
          </div>
          <div className="flex justify-between text-xs">
            <span className="text-green-600">Index: {formatNumber(tables?.index_scans ?? 0)}</span>
            <span className="text-yellow-600">Seq: {formatNumber(tables?.sequential_scans ?? 0)}</span>
          </div>
        </div>

        {bloatRatio > 20 && (
          <div className="flex items-center gap-2 p-3 bg-yellow-50 dark:bg-yellow-900/20 rounded-lg text-yellow-700 dark:text-yellow-300">
            <AlertTriangle className="w-4 h-4" />
            <span className="text-sm">Consider running VACUUM to reduce dead tuples</span>
          </div>
        )}
      </div>
    </motion.div>
  );
}

function LockActivity({ locks }: { locks: LockStats }) {
  const total = locks?.total ?? 0;
  const granted = locks?.granted ?? 0;
  const waiting = locks?.waiting ?? 0;
  const deadlocks = locks?.deadlocks ?? 0;

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: 0.4 }}
      className="bg-white dark:bg-gray-800 rounded-xl p-6 border border-gray-200 dark:border-gray-700 shadow-sm"
    >
      <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
        <Lock className="w-5 h-5 text-gray-400" />
        Lock Activity
      </h3>

      <div className="space-y-4">
        <div className="grid grid-cols-3 gap-4">
          <div className="text-center p-3 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
            <p className="text-2xl font-bold text-gray-900 dark:text-white">{total}</p>
            <p className="text-xs text-gray-500 dark:text-gray-400">Total Locks</p>
          </div>
          <div className="text-center p-3 bg-green-50 dark:bg-green-900/20 rounded-lg">
            <p className="text-2xl font-bold text-green-600">{granted}</p>
            <p className="text-xs text-green-600 dark:text-green-400">Granted</p>
          </div>
          <div className="text-center p-3 bg-red-50 dark:bg-red-900/20 rounded-lg">
            <p className="text-2xl font-bold text-red-600">{waiting}</p>
            <p className="text-xs text-red-600 dark:text-red-400">Waiting</p>
          </div>
        </div>

        <div className="pt-4 border-t border-gray-200 dark:border-gray-700">
          <div className="flex items-center justify-between">
            <span className="text-sm text-gray-500 dark:text-gray-400">Deadlocks</span>
            <span className={`text-lg font-semibold ${deadlocks > 0 ? 'text-red-600' : 'text-green-600'}`}>
              {deadlocks}
              {deadlocks === 0 && <CheckCircle className="w-4 h-4 inline ml-2" />}
            </span>
          </div>
        </div>

        {waiting > 5 && (
          <div className="flex items-center gap-2 p-3 bg-red-50 dark:bg-red-900/20 rounded-lg text-red-700 dark:text-red-300">
            <AlertTriangle className="w-4 h-4" />
            <span className="text-sm">High number of waiting locks detected</span>
          </div>
        )}
      </div>
    </motion.div>
  );
}

export default function PostgresStatsPage() {
  const { stats, allStats, loading, error, refresh, runVacuum, runAnalyze } = usePostgresStats();
  const [actionLoading, setActionLoading] = useState(false);
  const [selectedDatabaseId, setSelectedDatabaseId] = useState<string | null>(null);

  // Get all available databases
  const availableDatabases = Object.values(allStats);
  const databaseCount = availableDatabases.length;

  // Determine current stats: use selected, then stats from hook, then first from allStats
  const currentStats = selectedDatabaseId 
    ? allStats[selectedDatabaseId] 
    : stats || Object.values(allStats)[0];

  // Check if stats are actually populated (not just empty/zero values)
  const hasRealStats = currentStats && (
    currentStats.size_bytes > 0 ||
    (currentStats.tables && currentStats.tables.total_tables > 0) ||
    (currentStats.connections && currentStats.connections.total > 0) ||
    currentStats.collected_at
  );

  // Debug: Log the data structure
  useEffect(() => {
    if (databaseCount > 0) {
      console.log('Database Count:', databaseCount);
      console.log('Current Stats:', currentStats);
      console.log('Has Real Stats:', hasRealStats);
      if (currentStats) {
        console.log('Stats Details:', {
          database_id: currentStats.database_id,
          database_name: currentStats.database_name,
          size_bytes: currentStats.size_bytes,
          collected_at: currentStats.collected_at,
          tables: currentStats.tables,
          connections: currentStats.connections,
        });
      }
    }
  }, [allStats, currentStats, databaseCount, hasRealStats]);

  // Set initial selection when data loads
  useEffect(() => {
    if (!selectedDatabaseId && databaseCount > 0) {
      const firstDb = stats || Object.values(allStats)[0];
      if (firstDb) {
        setSelectedDatabaseId(firstDb.database_id);
      }
    }
    // Reset selection if selected database no longer exists
    if (selectedDatabaseId && !allStats[selectedDatabaseId] && databaseCount > 0) {
      const firstDb = Object.values(allStats)[0];
      if (firstDb) {
        setSelectedDatabaseId(firstDb.database_id);
      }
    }
  }, [stats, allStats, selectedDatabaseId, databaseCount]);

  const handleVacuum = async () => {
    if (!currentStats) return;
    setActionLoading(true);
    try {
      await runVacuum(currentStats.database_id);
      alert('Vacuum completed successfully');
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Vacuum failed');
    } finally {
      setActionLoading(false);
    }
  };

  const handleAnalyze = async () => {
    if (!currentStats) return;
    setActionLoading(true);
    try {
      await runAnalyze(currentStats.database_id);
      alert('Analyze completed successfully');
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Analyze failed');
    } finally {
      setActionLoading(false);
    }
  };

  if (loading && !currentStats) {
    return (
      <div className="flex items-center justify-center h-96">
        <RefreshCw className="w-8 h-8 animate-spin text-primary-500" />
      </div>
    );
  }

  if (!currentStats) {
    return (
      <div className="flex flex-col items-center justify-center h-96 text-gray-500 dark:text-gray-400">
        <Database className="w-12 h-12 mb-4 opacity-50" />
        <p className="text-lg font-medium">No PostgreSQL statistics available</p>
        <p className="text-sm mt-2">Connect a database to see statistics</p>
        {error && <p className="text-sm text-red-500 mt-2">{error}</p>}
        <button
          onClick={refresh}
          className="mt-4 px-4 py-2 bg-primary-500 text-white rounded-lg hover:bg-primary-600"
        >
          Refresh
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex-1">
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white flex items-center gap-3">
            <Database className="w-8 h-8 text-primary-500" />
            PostgreSQL Statistics
          </h1>
          <p className="text-gray-500 dark:text-gray-400 mt-1">
            {databaseCount > 0 ? `${databaseCount} database${databaseCount !== 1 ? 's' : ''} available` : 'No databases available'}
          </p>
        </div>
        <div className="flex items-center gap-3">
          {/* Database Selector */}
          {databaseCount > 1 && (
            <div className="relative">
              <label htmlFor="database-select" className="sr-only">
                Select Database
              </label>
              <select
                id="database-select"
                value={selectedDatabaseId || ''}
                onChange={(e) => setSelectedDatabaseId(e.target.value)}
                className="appearance-none bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 text-gray-900 dark:text-white rounded-lg px-4 py-2 pr-10 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent cursor-pointer min-w-[200px]"
              >
                {availableDatabases.map((db) => (
                  <option key={db.database_id} value={db.database_id}>
                    {db.database_name} ({formatBytes(db.size_bytes ?? 0)})
                  </option>
                ))}
              </select>
              <ChevronDown className="absolute right-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-500 pointer-events-none" />
            </div>
          )}
          {databaseCount === 1 && currentStats && (
            <div className="px-4 py-2 bg-gray-100 dark:bg-gray-700 rounded-lg text-sm text-gray-700 dark:text-gray-300">
              {currentStats.database_name}
            </div>
          )}
        </div>
      </div>

      {/* Current Database Info */}
      {currentStats && (
        <div className={`border rounded-lg p-4 ${
          hasRealStats 
            ? 'bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800' 
            : 'bg-yellow-50 dark:bg-yellow-900/20 border-yellow-200 dark:border-yellow-800'
        }`}>
          <div className="flex items-center justify-between">
            <div className="flex-1">
              <div className="flex items-center gap-2">
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                  {currentStats.database_name || currentStats.database_id || 'Unknown Database'}
                </h3>
                {!hasRealStats && (
                  <span className="px-2 py-1 text-xs bg-yellow-200 dark:bg-yellow-800 text-yellow-800 dark:text-yellow-200 rounded">
                    Collecting...
                  </span>
                )}
              </div>
              <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                Database ID: {currentStats.database_id || 'N/A'} • Size: {formatBytes(currentStats.size_bytes ?? 0)}
                {currentStats.collected_at && (
                  <span> • Last updated: {new Date(currentStats.collected_at).toLocaleString()}</span>
                )}
                {!currentStats.collected_at && (
                  <span> • Stats collection in progress...</span>
                )}
              </p>
              {!hasRealStats && (
                <p className="text-sm text-yellow-700 dark:text-yellow-300 mt-2 flex items-center gap-2">
                  <AlertTriangle className="w-4 h-4" />
                  Statistics are being collected. Please wait a moment and refresh.
                </p>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Action Buttons */}
      <div className="flex items-center justify-end gap-3">
        <button
          onClick={handleAnalyze}
          disabled={actionLoading || !currentStats}
          className="flex items-center gap-2 px-4 py-2 border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50"
        >
          <BarChart3 className="w-4 h-4" />
          Analyze
        </button>
        <button
          onClick={handleVacuum}
          disabled={actionLoading || !currentStats}
          className="flex items-center gap-2 px-4 py-2 border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50"
        >
          <HardDrive className="w-4 h-4" />
          Vacuum
        </button>
        <button
          onClick={refresh}
          disabled={loading}
          className="flex items-center gap-2 px-4 py-2 bg-primary-500 text-white rounded-lg hover:bg-primary-600 disabled:opacity-50"
        >
          <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
          Refresh
        </button>
      </div>

      {error && (
        <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4 flex items-center gap-3">
          <AlertTriangle className="w-5 h-5 text-red-500" />
          <span className="text-red-700 dark:text-red-300">{error}</span>
        </div>
      )}

      {/* Warning when stats are empty */}
      {currentStats && !hasRealStats && (
        <div className="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg p-4">
          <div className="flex items-start gap-3">
            <AlertTriangle className="w-5 h-5 text-yellow-600 dark:text-yellow-400 mt-0.5" />
            <div className="flex-1">
              <h4 className="font-semibold text-yellow-800 dark:text-yellow-200 mb-1">
                Statistics Not Available Yet
              </h4>
              <p className="text-sm text-yellow-700 dark:text-yellow-300">
                PostgreSQL statistics for this database are still being collected. This usually takes a few seconds after the database is registered.
                {databaseCount > 1 && (
                  <span> You can try selecting a different database from the dropdown above, or wait a moment and click Refresh.</span>
                )}
              </p>
              <button
                onClick={refresh}
                disabled={loading}
                className="mt-3 px-4 py-2 bg-yellow-600 text-white rounded-lg hover:bg-yellow-700 disabled:opacity-50 text-sm flex items-center gap-2"
              >
                <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
                Refresh Statistics
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Overview Stats - Only show if we have real stats */}
      {hasRealStats && (
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <StatCard
          title="Database Size"
          value={formatBytes(currentStats?.size_bytes ?? 0)}
          icon={HardDrive}
          color="bg-blue-100 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400"
        />
        <StatCard
          title="Total Tables"
          value={currentStats.tables?.total_tables ?? 0}
          subtitle={`${formatNumber(currentStats.tables?.total_rows ?? 0)} rows`}
          icon={Table}
          color="bg-purple-100 dark:bg-purple-900/30 text-purple-600 dark:text-purple-400"
        />
        <StatCard
          title="Index Hit Ratio"
          value={`${(currentStats.indexes?.index_hit_ratio ?? 0).toFixed(1)}%`}
          icon={Gauge}
          color="bg-green-100 dark:bg-green-900/30 text-green-600 dark:text-green-400"
          trend={(currentStats.indexes?.index_hit_ratio ?? 0) >= 99 ? 'up' : 'down'}
        />
        <StatCard
          title="Replicas"
          value={currentStats.replication?.replica_count ?? 0}
          subtitle={currentStats.replication?.is_replica ? 'This is a replica' : 'Primary node'}
          icon={Activity}
          color="bg-amber-100 dark:bg-amber-900/30 text-amber-600 dark:text-amber-400"
        />
      </div>
      )}

      {/* Detailed Stats - Only show if we have real stats */}
      {hasRealStats && (
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {currentStats.connections && <ConnectionsChart connections={currentStats.connections} />}
        {currentStats.queries && <QueryPerformance queries={currentStats.queries} />}
        {currentStats.tables && <TableOverview tables={currentStats.tables} />}
        {currentStats.locks && <LockActivity locks={currentStats.locks} />}
      </div>
      )}

      {/* Top Queries - Only show if we have real stats */}
      {hasRealStats && currentStats.queries?.top_queries && currentStats.queries.top_queries.length > 0 && (
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.5 }}
          className="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm"
        >
          <div className="p-6 border-b border-gray-200 dark:border-gray-700">
            <h2 className="text-lg font-semibold text-gray-900 dark:text-white flex items-center gap-2">
              <Clock className="w-5 h-5 text-gray-400" />
              Top Queries by Time
            </h2>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50 dark:bg-gray-700/50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Query</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Calls</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Total Time</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Avg Time</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Rows</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                {currentStats.queries.top_queries.map((query, i) => (
                  <tr key={i} className="hover:bg-gray-50 dark:hover:bg-gray-700/50">
                    <td className="px-6 py-4">
                      <code className="text-sm text-gray-600 dark:text-gray-300 font-mono truncate block max-w-md">
                        {query.query.substring(0, 60)}...
                      </code>
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-600 dark:text-gray-300">
                      {formatNumber(query.calls)}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-600 dark:text-gray-300">
                      {query.total_time_ms.toFixed(2)}ms
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-600 dark:text-gray-300">
                      {query.mean_time_ms.toFixed(2)}ms
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-600 dark:text-gray-300">
                      {formatNumber(query.rows)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </motion.div>
      )}
    </div>
  );
}

