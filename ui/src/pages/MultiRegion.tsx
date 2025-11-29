import { useState } from 'react';
import { motion } from 'framer-motion';
import {
  Globe,
  Activity,
  AlertTriangle,
  CheckCircle,
  XCircle,
  RefreshCw,
  ArrowRightLeft,
  Clock,
  Zap,
  MapPin,
  Signal,
} from 'lucide-react';
import { useMultiRegion } from '@/features/multiregion';
import type { RegionStatus } from '@/features/multiregion';
import Modal from '@/components/ui/Modal';

const regionCoordinates: Record<string, { x: number; y: number }> = {
  'us-west-2': { x: 15, y: 40 },
  'us-west-1': { x: 12, y: 38 },
  'us-east-1': { x: 28, y: 40 },
  'us-east-2': { x: 26, y: 42 },
  'eu-west-1': { x: 48, y: 32 },
  'eu-central-1': { x: 52, y: 30 },
  'ap-southeast-1': { x: 75, y: 55 },
  'ap-northeast-1': { x: 85, y: 38 },
};

export default function MultiRegion() {
  const { state, regions, replicationLag, loading, error, refresh, initiateFailover } = useMultiRegion();
  const [failoverModal, setFailoverModal] = useState(false);
  const [selectedRegion, setSelectedRegion] = useState<string | null>(null);
  const [failoverReason, setFailoverReason] = useState('');

  const handleFailover = async () => {
    if (!selectedRegion) return;
    try {
      await initiateFailover(selectedRegion, failoverReason || 'Manual failover');
      setFailoverModal(false);
      setSelectedRegion(null);
      setFailoverReason('');
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failover failed');
    }
  };

  const getStatusColor = (region: RegionStatus) => {
    if (!region.is_healthy) return 'bg-red-500';
    if (region.is_primary) return 'bg-emerald-500';
    return 'bg-blue-500';
  };

  const getStatusIcon = (region: RegionStatus) => {
    if (!region.is_healthy) return <XCircle className="w-4 h-4" />;
    if (region.is_primary) return <CheckCircle className="w-4 h-4" />;
    return <Activity className="w-4 h-4" />;
  };

  if (loading && !state) {
    return (
      <div className="flex items-center justify-center h-96">
        <RefreshCw className="w-8 h-8 animate-spin text-primary-500" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white flex items-center gap-3">
            <Globe className="w-8 h-8 text-primary-500" />
            Multi-Region Dashboard
          </h1>
          <p className="text-gray-500 dark:text-gray-400 mt-1">
            Global infrastructure status and cross-region coordination
          </p>
        </div>
        <button
          onClick={refresh}
          className="flex items-center gap-2 px-4 py-2 bg-primary-500 text-white rounded-lg hover:bg-primary-600 transition-colors"
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

      {/* Global Status Overview */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="bg-white dark:bg-gray-800 rounded-xl p-6 border border-gray-200 dark:border-gray-700 shadow-sm"
        >
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500 dark:text-gray-400">Primary Region</p>
              <p className="text-xl font-bold text-gray-900 dark:text-white mt-1">
                {state?.primary_region || 'N/A'}
              </p>
            </div>
            <div className="p-3 bg-emerald-100 dark:bg-emerald-900/30 rounded-lg">
              <MapPin className="w-6 h-6 text-emerald-600 dark:text-emerald-400" />
            </div>
          </div>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1 }}
          className="bg-white dark:bg-gray-800 rounded-xl p-6 border border-gray-200 dark:border-gray-700 shadow-sm"
        >
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500 dark:text-gray-400">Total Regions</p>
              <p className="text-xl font-bold text-gray-900 dark:text-white mt-1">
                {regions.length}
              </p>
            </div>
            <div className="p-3 bg-blue-100 dark:bg-blue-900/30 rounded-lg">
              <Globe className="w-6 h-6 text-blue-600 dark:text-blue-400" />
            </div>
          </div>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2 }}
          className="bg-white dark:bg-gray-800 rounded-xl p-6 border border-gray-200 dark:border-gray-700 shadow-sm"
        >
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500 dark:text-gray-400">Healthy Regions</p>
              <p className="text-xl font-bold text-gray-900 dark:text-white mt-1">
                {regions.filter(r => r.is_healthy).length} / {regions.length}
              </p>
            </div>
            <div className="p-3 bg-green-100 dark:bg-green-900/30 rounded-lg">
              <CheckCircle className="w-6 h-6 text-green-600 dark:text-green-400" />
            </div>
          </div>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.3 }}
          className="bg-white dark:bg-gray-800 rounded-xl p-6 border border-gray-200 dark:border-gray-700 shadow-sm"
        >
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500 dark:text-gray-400">Failover Status</p>
              <p className="text-xl font-bold text-gray-900 dark:text-white mt-1">
                {state?.failover_enabled ? 'Enabled' : 'Disabled'}
              </p>
            </div>
            <div className={`p-3 rounded-lg ${state?.failover_enabled ? 'bg-emerald-100 dark:bg-emerald-900/30' : 'bg-gray-100 dark:bg-gray-700'}`}>
              <ArrowRightLeft className={`w-6 h-6 ${state?.failover_enabled ? 'text-emerald-600 dark:text-emerald-400' : 'text-gray-500'}`} />
            </div>
          </div>
        </motion.div>
      </div>

      {/* World Map Visualization */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.4 }}
        className="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm overflow-hidden"
      >
        <div className="p-6 border-b border-gray-200 dark:border-gray-700">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">Global Region Map</h2>
        </div>
        <div className="relative h-80 bg-gradient-to-br from-slate-900 via-blue-900 to-slate-900 overflow-hidden">
          {/* Grid pattern */}
          <div className="absolute inset-0 opacity-10">
            <svg className="w-full h-full">
              <defs>
                <pattern id="grid" width="40" height="40" patternUnits="userSpaceOnUse">
                  <path d="M 40 0 L 0 0 0 40" fill="none" stroke="white" strokeWidth="0.5"/>
                </pattern>
              </defs>
              <rect width="100%" height="100%" fill="url(#grid)" />
            </svg>
          </div>

          {/* Region dots */}
          {regions.map((region, index) => {
            const coords = regionCoordinates[region.name] || { x: 50, y: 50 };
            return (
              <motion.div
                key={region.name}
                initial={{ scale: 0 }}
                animate={{ scale: 1 }}
                transition={{ delay: 0.5 + index * 0.1 }}
                className="absolute transform -translate-x-1/2 -translate-y-1/2 cursor-pointer group"
                style={{ left: `${coords.x}%`, top: `${coords.y}%` }}
                onClick={() => {
                  if (!region.is_primary && region.is_healthy) {
                    setSelectedRegion(region.name);
                    setFailoverModal(true);
                  }
                }}
              >
                {/* Pulse animation for primary */}
                {region.is_primary && (
                  <div className="absolute inset-0 -m-2 animate-ping">
                    <div className="w-8 h-8 rounded-full bg-emerald-500/30" />
                  </div>
                )}
                
                {/* Region dot */}
                <div className={`relative w-4 h-4 rounded-full ${getStatusColor(region)} shadow-lg`}>
                  <div className="absolute inset-0 rounded-full animate-pulse opacity-50" style={{ backgroundColor: 'inherit' }} />
                </div>

                {/* Tooltip */}
                <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-10">
                  <div className="bg-gray-900 text-white text-xs rounded-lg px-3 py-2 whitespace-nowrap shadow-xl">
                    <div className="font-semibold">{region.name}</div>
                    <div className="flex items-center gap-1 mt-1">
                      {getStatusIcon(region)}
                      <span>{region.is_primary ? 'Primary' : region.is_healthy ? 'Healthy' : 'Unhealthy'}</span>
                    </div>
                    <div className="text-gray-400 mt-1">Latency: {region.latency}ms</div>
                  </div>
                  <div className="absolute top-full left-1/2 -translate-x-1/2 border-4 border-transparent border-t-gray-900" />
                </div>
              </motion.div>
            );
          })}

          {/* Connection lines */}
          <svg className="absolute inset-0 w-full h-full pointer-events-none">
            {regions.filter(r => !r.is_primary).map((region) => {
              const primary = regions.find(r => r.is_primary);
              if (!primary) return null;
              const fromCoords = regionCoordinates[primary.name] || { x: 50, y: 50 };
              const toCoords = regionCoordinates[region.name] || { x: 50, y: 50 };
              return (
                <line
                  key={`line-${region.name}`}
                  x1={`${fromCoords.x}%`}
                  y1={`${fromCoords.y}%`}
                  x2={`${toCoords.x}%`}
                  y2={`${toCoords.y}%`}
                  stroke={region.is_healthy ? '#10b981' : '#ef4444'}
                  strokeWidth="1"
                  strokeDasharray="4,4"
                  opacity="0.5"
                />
              );
            })}
          </svg>
        </div>
      </motion.div>

      {/* Region Details */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Region List */}
        <motion.div
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ delay: 0.5 }}
          className="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm"
        >
          <div className="p-6 border-b border-gray-200 dark:border-gray-700">
            <h2 className="text-lg font-semibold text-gray-900 dark:text-white">Region Status</h2>
          </div>
          <div className="divide-y divide-gray-200 dark:divide-gray-700">
            {regions.map((region) => (
              <div key={region.name} className="p-4 flex items-center justify-between hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors">
                <div className="flex items-center gap-4">
                  <div className={`w-3 h-3 rounded-full ${getStatusColor(region)}`} />
                  <div>
                    <div className="font-medium text-gray-900 dark:text-white flex items-center gap-2">
                      {region.name}
                      {region.is_primary && (
                        <span className="text-xs bg-emerald-100 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400 px-2 py-0.5 rounded-full">
                          Primary
                        </span>
                      )}
                    </div>
                    <div className="text-sm text-gray-500 dark:text-gray-400">{region.endpoint}</div>
                  </div>
                </div>
                <div className="flex items-center gap-4">
                  <div className="text-right">
                    <div className="text-sm font-medium text-gray-900 dark:text-white flex items-center gap-1">
                      <Signal className="w-4 h-4 text-gray-400" />
                      {region.latency}ms
                    </div>
                    <div className="text-xs text-gray-500 dark:text-gray-400">
                      {region.active_connections} connections
                    </div>
                  </div>
                  {!region.is_primary && region.is_healthy && (
                    <button
                      onClick={() => {
                        setSelectedRegion(region.name);
                        setFailoverModal(true);
                      }}
                      className="px-3 py-1.5 text-sm bg-orange-100 dark:bg-orange-900/30 text-orange-700 dark:text-orange-400 rounded-lg hover:bg-orange-200 dark:hover:bg-orange-900/50 transition-colors"
                    >
                      Failover
                    </button>
                  )}
                </div>
              </div>
            ))}
          </div>
        </motion.div>

        {/* Replication Lag */}
        <motion.div
          initial={{ opacity: 0, x: 20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ delay: 0.6 }}
          className="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm"
        >
          <div className="p-6 border-b border-gray-200 dark:border-gray-700">
            <h2 className="text-lg font-semibold text-gray-900 dark:text-white flex items-center gap-2">
              <Clock className="w-5 h-5 text-gray-400" />
              Replication Lag
            </h2>
          </div>
          <div className="p-6 space-y-4">
            {replicationLag.length === 0 ? (
              <div className="text-center py-8 text-gray-500 dark:text-gray-400">
                <Zap className="w-12 h-12 mx-auto mb-3 opacity-50" />
                <p>No replication lag data available</p>
              </div>
            ) : (
              replicationLag.map((lag) => (
                <div key={lag.region} className="space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium text-gray-900 dark:text-white">{lag.region}</span>
                    <span className={`text-sm font-medium ${
                      lag.status === 'synced' ? 'text-green-600' : 
                      lag.status === 'lagging' ? 'text-yellow-600' : 'text-red-600'
                    }`}>
                      {lag.lag_ms}ms
                    </span>
                  </div>
                  <div className="h-2 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
                    <div
                      className={`h-full rounded-full transition-all ${
                        lag.status === 'synced' ? 'bg-green-500' :
                        lag.status === 'lagging' ? 'bg-yellow-500' : 'bg-red-500'
                      }`}
                      style={{ width: `${Math.min(100, (lag.lag_ms / 1000) * 100)}%` }}
                    />
                  </div>
                  <div className="text-xs text-gray-500 dark:text-gray-400">
                    Last sync: {new Date(lag.last_sync).toLocaleTimeString()}
                  </div>
                </div>
              ))
            )}
          </div>
        </motion.div>
      </div>

      {/* Failover Modal */}
      <Modal
        isOpen={failoverModal}
        onClose={() => {
          setFailoverModal(false);
          setSelectedRegion(null);
          setFailoverReason('');
        }}
        title="Initiate Failover"
      >
        <div className="space-y-4">
          <div className="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg p-4">
            <div className="flex items-start gap-3">
              <AlertTriangle className="w-5 h-5 text-yellow-600 dark:text-yellow-400 mt-0.5" />
              <div>
                <p className="text-sm font-medium text-yellow-800 dark:text-yellow-200">Warning</p>
                <p className="text-sm text-yellow-700 dark:text-yellow-300 mt-1">
                  Failover will switch the primary region to <strong>{selectedRegion}</strong>. 
                  This may cause brief interruption to services.
                </p>
              </div>
            </div>
          </div>
          
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Reason for failover
            </label>
            <textarea
              value={failoverReason}
              onChange={(e) => setFailoverReason(e.target.value)}
              placeholder="Enter reason for initiating failover..."
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              rows={3}
            />
          </div>

          <div className="flex gap-3">
            <button
              onClick={() => {
                setFailoverModal(false);
                setSelectedRegion(null);
              }}
              className="flex-1 px-4 py-2 border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
            >
              Cancel
            </button>
            <button
              onClick={handleFailover}
              className="flex-1 px-4 py-2 bg-orange-500 text-white rounded-lg hover:bg-orange-600 transition-colors"
            >
              Confirm Failover
            </button>
          </div>
        </div>
      </Modal>
    </div>
  );
}

