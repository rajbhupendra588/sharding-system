import { useState } from 'react';
import { motion } from 'framer-motion';
import {
  Shield,
  AlertTriangle,
  CheckCircle,
  XCircle,
  RefreshCw,
  ArrowRightLeft,
  Clock,
  Target,
  Play,
  History,
  Activity,
  Zap,
  ArrowRight,
} from 'lucide-react';
import { useDisasterRecovery } from '@/features/disaster-recovery';
import type { FailoverEvent, DrillResult } from '@/features/disaster-recovery';
import Modal from '@/components/ui/Modal';

export default function DisasterRecovery() {
  const { status, loading, error, actionInProgress, refresh, failover, failback, runDrill } = useDisasterRecovery();
  const [failoverModal, setFailoverModal] = useState(false);
  const [drillModal, setDrillModal] = useState(false);
  const [selectedRegion, setSelectedRegion] = useState('');
  const [failoverReason, setFailoverReason] = useState('');
  const [drillResult, setDrillResult] = useState<DrillResult | null>(null);

  const handleFailover = async () => {
    if (!selectedRegion) return;
    try {
      await failover(selectedRegion, failoverReason || 'Manual failover');
      setFailoverModal(false);
      setSelectedRegion('');
      setFailoverReason('');
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failover failed');
    }
  };

  const handleFailback = async () => {
    if (!confirm('Are you sure you want to failback to the original primary region?')) return;
    try {
      await failback();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failback failed');
    }
  };

  const handleRunDrill = async () => {
    if (!selectedRegion) return;
    try {
      const result = await runDrill(selectedRegion);
      setDrillResult(result);
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Drill failed');
    }
  };

  const formatDuration = (ms: number) => {
    if (ms < 1000) return `${ms}ms`;
    if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
    return `${(ms / 60000).toFixed(1)}m`;
  };

  if (loading && !status) {
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
            <Shield className="w-8 h-8 text-primary-500" />
            Disaster Recovery
          </h1>
          <p className="text-gray-500 dark:text-gray-400 mt-1">
            Manage failover, failback, and run DR drills
          </p>
        </div>
        <div className="flex gap-3">
          <button
            onClick={refresh}
            disabled={loading}
            className="flex items-center gap-2 px-4 py-2 border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
          >
            <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </button>
          <button
            onClick={() => setDrillModal(true)}
            className="flex items-center gap-2 px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"
          >
            <Play className="w-4 h-4" />
            Run DR Drill
          </button>
        </div>
      </div>

      {error && (
        <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4 flex items-center gap-3">
          <AlertTriangle className="w-5 h-5 text-red-500" />
          <span className="text-red-700 dark:text-red-300">{error}</span>
        </div>
      )}

      {/* Current Status Banner */}
      {status?.is_failed_over && (
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          className="bg-orange-50 dark:bg-orange-900/20 border-2 border-orange-300 dark:border-orange-700 rounded-xl p-6"
        >
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <div className="p-3 bg-orange-100 dark:bg-orange-900/50 rounded-full">
                <AlertTriangle className="w-8 h-8 text-orange-600 dark:text-orange-400" />
              </div>
              <div>
                <h3 className="text-lg font-semibold text-orange-800 dark:text-orange-200">
                  System is in Failover State
                </h3>
                <p className="text-orange-700 dark:text-orange-300">
                  Currently running on <strong>{status.current_region}</strong> instead of primary <strong>{status.primary_region}</strong>
                </p>
              </div>
            </div>
            <button
              onClick={handleFailback}
              disabled={actionInProgress}
              className="flex items-center gap-2 px-4 py-2 bg-orange-600 text-white rounded-lg hover:bg-orange-700 transition-colors disabled:opacity-50"
            >
              <ArrowRightLeft className="w-4 h-4" />
              Failback to Primary
            </button>
          </div>
        </motion.div>
      )}

      {/* DR Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="bg-white dark:bg-gray-800 rounded-xl p-6 border border-gray-200 dark:border-gray-700 shadow-sm"
        >
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500 dark:text-gray-400">RPO (Max Data Loss)</p>
              <p className="text-2xl font-bold text-gray-900 dark:text-white mt-1">
                {status?.rpo || 'N/A'}
              </p>
            </div>
            <div className="p-3 bg-purple-100 dark:bg-purple-900/30 rounded-lg">
              <Target className="w-6 h-6 text-purple-600 dark:text-purple-400" />
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
              <p className="text-sm text-gray-500 dark:text-gray-400">RTO (Max Downtime)</p>
              <p className="text-2xl font-bold text-gray-900 dark:text-white mt-1">
                {status?.rto || 'N/A'}
              </p>
            </div>
            <div className="p-3 bg-blue-100 dark:bg-blue-900/30 rounded-lg">
              <Clock className="w-6 h-6 text-blue-600 dark:text-blue-400" />
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
              <p className="text-sm text-gray-500 dark:text-gray-400">Auto Failover</p>
              <p className="text-2xl font-bold text-gray-900 dark:text-white mt-1">
                {status?.auto_failover ? 'Enabled' : 'Disabled'}
              </p>
            </div>
            <div className={`p-3 rounded-lg ${status?.auto_failover ? 'bg-green-100 dark:bg-green-900/30' : 'bg-gray-100 dark:bg-gray-700'}`}>
              <Zap className={`w-6 h-6 ${status?.auto_failover ? 'text-green-600 dark:text-green-400' : 'text-gray-500'}`} />
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
              <p className="text-sm text-gray-500 dark:text-gray-400">Failover Events</p>
              <p className="text-2xl font-bold text-gray-900 dark:text-white mt-1">
                {status?.failover_history?.length || 0}
              </p>
            </div>
            <div className="p-3 bg-amber-100 dark:bg-amber-900/30 rounded-lg">
              <History className="w-6 h-6 text-amber-600 dark:text-amber-400" />
            </div>
          </div>
        </motion.div>
      </div>

      {/* Region Health */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.4 }}
        className="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm"
      >
        <div className="p-6 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white flex items-center gap-2">
            <Activity className="w-5 h-5 text-gray-400" />
            Region Health Status
          </h2>
          <button
            onClick={() => setFailoverModal(true)}
            disabled={!status || status.is_failed_over}
            className="px-4 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
          >
            <ArrowRightLeft className="w-4 h-4" />
            Manual Failover
          </button>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50 dark:bg-gray-700/50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Region</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Status</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Latency</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Replication Lag</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Last Check</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Consecutive Fails</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
              {status?.region_statuses?.map((region) => (
                <tr key={region.region} className="hover:bg-gray-50 dark:hover:bg-gray-700/50">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center gap-3">
                      <span className="font-medium text-gray-900 dark:text-white">{region.region}</span>
                      {region.region === status.primary_region && (
                        <span className="text-xs bg-emerald-100 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400 px-2 py-0.5 rounded-full">
                          Primary
                        </span>
                      )}
                      {region.region === status.current_region && region.region !== status.primary_region && (
                        <span className="text-xs bg-orange-100 dark:bg-orange-900/30 text-orange-700 dark:text-orange-400 px-2 py-0.5 rounded-full">
                          Active
                        </span>
                      )}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium ${
                      region.is_healthy 
                        ? 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400'
                        : 'bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400'
                    }`}>
                      {region.is_healthy ? <CheckCircle className="w-3.5 h-3.5" /> : <XCircle className="w-3.5 h-3.5" />}
                      {region.is_healthy ? 'Healthy' : 'Unhealthy'}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600 dark:text-gray-300">
                    {region.latency}ms
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600 dark:text-gray-300">
                    {formatDuration(region.replication_lag)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600 dark:text-gray-300">
                    {new Date(region.last_check).toLocaleTimeString()}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className={`text-sm font-medium ${region.consecutive_fails > 0 ? 'text-red-600' : 'text-gray-600 dark:text-gray-300'}`}>
                      {region.consecutive_fails}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </motion.div>

      {/* Failover History */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.5 }}
        className="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm"
      >
        <div className="p-6 border-b border-gray-200 dark:border-gray-700">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white flex items-center gap-2">
            <History className="w-5 h-5 text-gray-400" />
            Failover History
          </h2>
        </div>
        <div className="divide-y divide-gray-200 dark:divide-gray-700">
          {(!status?.failover_history || status.failover_history.length === 0) ? (
            <div className="p-8 text-center text-gray-500 dark:text-gray-400">
              <Shield className="w-12 h-12 mx-auto mb-3 opacity-50" />
              <p>No failover events recorded</p>
            </div>
          ) : (
            status.failover_history.slice(-10).reverse().map((event: FailoverEvent) => (
              <div key={event.id} className="p-4 flex items-center justify-between">
                <div className="flex items-center gap-4">
                  <div className={`p-2 rounded-full ${event.success ? 'bg-green-100 dark:bg-green-900/30' : 'bg-red-100 dark:bg-red-900/30'}`}>
                    {event.success ? (
                      <CheckCircle className="w-5 h-5 text-green-600 dark:text-green-400" />
                    ) : (
                      <XCircle className="w-5 h-5 text-red-600 dark:text-red-400" />
                    )}
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-gray-900 dark:text-white">{event.from_region}</span>
                      <ArrowRight className="w-4 h-4 text-gray-400" />
                      <span className="font-medium text-gray-900 dark:text-white">{event.to_region}</span>
                      {event.automatic && (
                        <span className="text-xs bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400 px-2 py-0.5 rounded-full">
                          Automatic
                        </span>
                      )}
                    </div>
                    <p className="text-sm text-gray-500 dark:text-gray-400">{event.reason}</p>
                  </div>
                </div>
                <div className="text-right">
                  <p className="text-sm text-gray-600 dark:text-gray-300">
                    {new Date(event.start_time).toLocaleString()}
                  </p>
                  {event.duration && (
                    <p className="text-xs text-gray-500 dark:text-gray-400">
                      Duration: {formatDuration(event.duration)}
                    </p>
                  )}
                </div>
              </div>
            ))
          )}
        </div>
      </motion.div>

      {/* Failover Modal */}
      <Modal
        isOpen={failoverModal}
        onClose={() => {
          setFailoverModal(false);
          setSelectedRegion('');
          setFailoverReason('');
        }}
        title="Manual Failover"
      >
        <div className="space-y-4">
          <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
            <div className="flex items-start gap-3">
              <AlertTriangle className="w-5 h-5 text-red-600 dark:text-red-400 mt-0.5" />
              <div>
                <p className="text-sm font-medium text-red-800 dark:text-red-200">Critical Action</p>
                <p className="text-sm text-red-700 dark:text-red-300 mt-1">
                  Manual failover will switch all traffic to the selected region. 
                  This may cause temporary service disruption.
                </p>
              </div>
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Target Region
            </label>
            <select
              value={selectedRegion}
              onChange={(e) => setSelectedRegion(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary-500"
            >
              <option value="">Select a region</option>
              {status?.region_statuses
                ?.filter(r => r.region !== status.primary_region && r.is_healthy)
                .map(r => (
                  <option key={r.region} value={r.region}>{r.region}</option>
                ))}
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Reason
            </label>
            <textarea
              value={failoverReason}
              onChange={(e) => setFailoverReason(e.target.value)}
              placeholder="Enter reason for failover..."
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary-500"
              rows={3}
            />
          </div>

          <div className="flex gap-3">
            <button
              onClick={() => setFailoverModal(false)}
              className="flex-1 px-4 py-2 border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700"
            >
              Cancel
            </button>
            <button
              onClick={handleFailover}
              disabled={!selectedRegion || actionInProgress}
              className="flex-1 px-4 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600 disabled:opacity-50"
            >
              {actionInProgress ? 'Processing...' : 'Initiate Failover'}
            </button>
          </div>
        </div>
      </Modal>

      {/* DR Drill Modal */}
      <Modal
        isOpen={drillModal}
        onClose={() => {
          setDrillModal(false);
          setSelectedRegion('');
          setDrillResult(null);
        }}
        title="Disaster Recovery Drill"
      >
        <div className="space-y-4">
          {!drillResult ? (
            <>
              <p className="text-gray-600 dark:text-gray-300">
                Run a DR drill to verify failover readiness without actually performing a failover.
              </p>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Target Region
                </label>
                <select
                  value={selectedRegion}
                  onChange={(e) => setSelectedRegion(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                >
                  <option value="">Select a region</option>
                  {status?.region_statuses
                    ?.filter(r => r.region !== status.current_region)
                    .map(r => (
                      <option key={r.region} value={r.region}>{r.region}</option>
                    ))}
                </select>
              </div>

              <div className="flex gap-3">
                <button
                  onClick={() => setDrillModal(false)}
                  className="flex-1 px-4 py-2 border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700"
                >
                  Cancel
                </button>
                <button
                  onClick={handleRunDrill}
                  disabled={!selectedRegion || actionInProgress}
                  className="flex-1 px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50"
                >
                  {actionInProgress ? 'Running...' : 'Run Drill'}
                </button>
              </div>
            </>
          ) : (
            <div className="space-y-4">
              <div className={`p-4 rounded-lg ${drillResult.all_passed ? 'bg-green-50 dark:bg-green-900/20' : 'bg-red-50 dark:bg-red-900/20'}`}>
                <div className="flex items-center gap-3">
                  {drillResult.all_passed ? (
                    <CheckCircle className="w-6 h-6 text-green-600" />
                  ) : (
                    <XCircle className="w-6 h-6 text-red-600" />
                  )}
                  <div>
                    <p className={`font-semibold ${drillResult.all_passed ? 'text-green-800 dark:text-green-200' : 'text-red-800 dark:text-red-200'}`}>
                      {drillResult.all_passed ? 'All Checks Passed' : 'Some Checks Failed'}
                    </p>
                    <p className="text-sm text-gray-600 dark:text-gray-400">
                      Duration: {formatDuration(drillResult.duration)}
                    </p>
                  </div>
                </div>
              </div>

              <div className="space-y-2">
                {drillResult.checks.map((check, i) => (
                  <div key={i} className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
                    <div className="flex items-center gap-3">
                      {check.passed ? (
                        <CheckCircle className="w-5 h-5 text-green-500" />
                      ) : (
                        <XCircle className="w-5 h-5 text-red-500" />
                      )}
                      <span className="font-medium text-gray-900 dark:text-white">{check.name}</span>
                    </div>
                    <span className="text-sm text-gray-500 dark:text-gray-400">{check.message}</span>
                  </div>
                ))}
              </div>

              <button
                onClick={() => {
                  setDrillModal(false);
                  setDrillResult(null);
                  setSelectedRegion('');
                }}
                className="w-full px-4 py-2 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600"
              >
                Close
              </button>
            </div>
          )}
        </div>
      </Modal>
    </div>
  );
}

