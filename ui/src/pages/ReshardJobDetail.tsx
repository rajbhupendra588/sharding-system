import { useParams, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { ArrowLeft, RefreshCw } from 'lucide-react';
import { apiClient } from '@/lib/api-client';
import LoadingSpinner from '@/components/ui/LoadingSpinner';
import Button from '@/components/ui/Button';
import StatusBadge from '@/components/ui/StatusBadge';
import { formatDate, formatDuration } from '@/lib/utils';

export default function ReshardJobDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const { data: job, isLoading, refetch } = useQuery({
    queryKey: ['reshard-job', id],
    queryFn: () => apiClient.getReshardJob(id!),
    enabled: !!id,
    refetchInterval: (query) => {
      // Auto-refresh if job is still in progress
      const job = query.state.data;
      if (job?.status && ['pending', 'precopy', 'deltasync', 'cutover', 'validation'].includes(job.status)) {
        return 5000; // Refresh every 5 seconds
      }
      return false; // Don't auto-refresh if completed or failed
    },
  });

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <LoadingSpinner size="lg" />
      </div>
    );
  }

  if (!job) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500">Job not found</p>
        <Button onClick={() => navigate('/resharding')} className="mt-4">
          Back to Resharding
        </Button>
      </div>
    );
  }

  const progressPercentage = Math.round(job.progress * 100);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <Button variant="secondary" onClick={() => navigate('/resharding')}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back
          </Button>
          <div>
            <h1 className="text-3xl font-bold text-gray-900 dark:text-white">Resharding Job</h1>
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">{job.id}</p>
          </div>
        </div>
        <div className="flex items-center space-x-2">
          <StatusBadge status={job.status} />
          <Button variant="secondary" onClick={() => refetch()}>
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh
          </Button>
        </div>
      </div>

      {/* Progress */}
      <div className="card">
        <div className="flex items-center justify-between mb-2">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">Progress</h2>
          <span className="text-sm font-medium text-gray-700 dark:text-gray-300">{progressPercentage}%</span>
        </div>
        <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-4 mb-4">
          <div
            className="bg-primary-600 h-4 rounded-full transition-all duration-300"
            style={{ width: `${progressPercentage}%` }}
          />
        </div>
        {job.total_keys > 0 && (
          <div className="text-sm text-gray-600 dark:text-gray-400">
            {job.keys_migrated.toLocaleString()} / {job.total_keys.toLocaleString()} keys migrated
          </div>
        )}
      </div>

      {/* Details Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Job Information */}
        <div className="card">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Job Information</h2>
          <dl className="space-y-3">
            <div>
              <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Type</dt>
              <dd className="mt-1 text-sm text-gray-900 dark:text-white capitalize">{job.type}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Status</dt>
              <dd className="mt-1">
                <StatusBadge status={job.status} />
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Started</dt>
              <dd className="mt-1 text-sm text-gray-900 dark:text-white">{formatDate(job.started_at)}</dd>
            </div>
            {job.completed_at && (
              <div>
                <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Completed</dt>
                <dd className="mt-1 text-sm text-gray-900 dark:text-white">{formatDate(job.completed_at)}</dd>
              </div>
            )}
            {job.completed_at && (
              <div>
                <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Duration</dt>
                <dd className="mt-1 text-sm text-gray-900 dark:text-white">
                  {formatDuration(
                    new Date(job.completed_at).getTime() - new Date(job.started_at).getTime()
                  )}
                </dd>
              </div>
            )}
          </dl>
        </div>

        {/* Shards */}
        <div className="card">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Shards</h2>
          <div className="space-y-4">
            <div>
              <dt className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">Source Shards</dt>
              <div className="space-y-1">
                {job.source_shards.map((shardId) => (
                  <div
                    key={shardId}
                    className="text-sm text-gray-900 bg-gray-50 px-3 py-2 rounded font-mono"
                  >
                    {shardId}
                  </div>
                ))}
              </div>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">Target Shards</dt>
              <div className="space-y-1">
                {job.target_shards.map((shardId) => (
                  <div
                    key={shardId}
                    className="text-sm text-gray-900 bg-gray-50 px-3 py-2 rounded font-mono"
                  >
                    {shardId}
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Error Message */}
      {job.error_message && (
        <div className="card bg-red-50 border-red-200 dark:bg-red-900/20 dark:border-red-800">
          <h2 className="text-lg font-semibold text-red-900 dark:text-red-300 mb-2">Error</h2>
          <p className="text-sm text-red-700 dark:text-red-400">{job.error_message}</p>
        </div>
      )}
    </div>
  );
}

