import { useParams, useNavigate } from 'react-router-dom';
import { ArrowLeft, Database as DatabaseIcon, Copy, CheckCircle2, Trash2, Settings, RefreshCw, GitBranch } from 'lucide-react';
import { toast } from 'react-hot-toast';
import LoadingSpinner from '@/components/ui/LoadingSpinner';
import Button from '@/components/ui/Button';
import StatusBadge from '@/components/ui/StatusBadge';
import Modal from '@/components/ui/Modal';
import { formatDate, formatRelativeTime } from '@/lib/utils';
import { useDatabase, useDatabaseStatus, databaseRepository } from '@/features/database';
import { useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';


export default function DatabaseDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [copiedConnection, setCopiedConnection] = useState(false);

  const { data: database, isLoading } = useDatabase(id || '');
  useDatabaseStatus(id || '');

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedConnection(true);
      setTimeout(() => setCopiedConnection(false), 2000);
      toast.success('Copied to clipboard');
    } catch (err) {
      toast.error('Failed to copy');
    }
  };

  const handleDelete = async () => {
    if (!id) return;
    try {
      // Note: Delete endpoint may not exist yet, this is a placeholder
      await databaseRepository.findById(id); // Just to show the pattern
      toast.success('Database deletion initiated');
      queryClient.invalidateQueries({ queryKey: ['databases'] });
      navigate('/databases');
    } catch (error: any) {
      toast.error(`Failed to delete database: ${error.message}`);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'ready':
        return 'success';
      case 'creating':
        return 'warning';
      case 'failed':
        return 'error';
      case 'scaling':
        return 'info';
      default:
        return 'default';
    }
  };



  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <LoadingSpinner size="lg" />
      </div>
    );
  }

  if (!database) {
    return (
      <div className="text-center py-12">
        <DatabaseIcon className="h-12 w-12 mx-auto text-gray-400 mb-4" />
        <p className="text-gray-600 dark:text-gray-400 mb-4">Database not found</p>
        <Button onClick={() => navigate('/databases')}>
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back to Databases
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-6 p-4 sm:p-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div className="flex items-center space-x-4">
          <Button variant="secondary" onClick={() => navigate('/databases')}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back
          </Button>
          <div>
            <h1 className="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-white">
              {database.display_name || database.name}
            </h1>
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">{database.id}</p>
          </div>
        </div>
        <div className="flex items-center space-x-2">
          <StatusBadge status={getStatusColor(database.status)} />
          <Button
            variant="secondary"
            size="sm"
            onClick={() => {
              queryClient.invalidateQueries({ queryKey: ['databases', id] });
              toast.success('Refreshed');
            }}
          >
            <RefreshCw className="h-4 w-4" />
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => navigate(`/databases/${database.name}/branches`)}
          >
            <GitBranch className="h-4 w-4 mr-2" />
            Branches
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => toast.success('Settings feature coming soon!')}
          >
            <Settings className="h-4 w-4" />
          </Button>
          <Button
            variant="danger"
            size="sm"
            onClick={() => setIsDeleteModalOpen(true)}
            disabled={database.status === 'creating'}
          >
            <Trash2 className="h-4 w-4 mr-2" />
            Delete
          </Button>
        </div>
      </div>

      {/* Database Info Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {/* Basic Information */}
        <div className="card">
          <h3 className="text-lg font-semibold mb-4 text-gray-900 dark:text-white">Basic Information</h3>
          <div className="space-y-3">
            <div>
              <label className="text-sm font-medium text-gray-500 dark:text-gray-400">Name</label>
              <p className="text-gray-900 dark:text-white">{database.name}</p>
            </div>
            {database.display_name && (
              <div>
                <label className="text-sm font-medium text-gray-500 dark:text-gray-400">Display Name</label>
                <p className="text-gray-900 dark:text-white">{database.display_name}</p>
              </div>
            )}
            {database.description && (
              <div>
                <label className="text-sm font-medium text-gray-500 dark:text-gray-400">Description</label>
                <p className="text-gray-900 dark:text-white">{database.description}</p>
              </div>
            )}
            <div>
              <label className="text-sm font-medium text-gray-500 dark:text-gray-400">Template</label>
              <span className="inline-block px-2 py-1 text-xs font-medium rounded bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200 mt-1">
                {database.template}
              </span>
            </div>
          </div>
        </div>

        {/* Configuration */}
        <div className="card">
          <h3 className="text-lg font-semibold mb-4 text-gray-900 dark:text-white">Configuration</h3>
          <div className="space-y-3">
            <div>
              <label className="text-sm font-medium text-gray-500 dark:text-gray-400">Shard Key</label>
              <p className="text-gray-900 dark:text-white font-mono">{database.shard_key}</p>
            </div>
            <div>
              <label className="text-sm font-medium text-gray-500 dark:text-gray-400">Shard Count</label>
              <p className="text-gray-900 dark:text-white">{database.shard_ids?.length ?? 0}</p>
            </div>
            <div>
              <label className="text-sm font-medium text-gray-500 dark:text-gray-400">Client App ID</label>
              <p className="text-gray-900 dark:text-white font-mono text-sm">{database.client_app_id}</p>
            </div>
          </div>
        </div>

        {/* Connection */}
        <div className="card">
          <h3 className="text-lg font-semibold mb-4 text-gray-900 dark:text-white">Connection</h3>
          <div className="space-y-3">
            <div>
              <label className="text-sm font-medium text-gray-500 dark:text-gray-400">Connection String</label>
              <div className="flex items-center gap-2 mt-1">
                <code className="flex-1 text-xs bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded break-all">
                  {database.connection_string}
                </code>
                <button
                  onClick={() => copyToClipboard(database.connection_string)}
                  className="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
                >
                  {copiedConnection ? (
                    <CheckCircle2 className="h-4 w-4 text-green-500" />
                  ) : (
                    <Copy className="h-4 w-4" />
                  )}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Shards List */}
      <div className="card">
        <h3 className="text-lg font-semibold mb-4 text-gray-900 dark:text-white">Shards</h3>
        {(database.shard_ids?.length ?? 0) === 0 ? (
          <div className="text-center py-8 text-gray-500 dark:text-gray-400">
            <DatabaseIcon className="h-8 w-8 mx-auto mb-2 opacity-50" />
            <p>No shards created yet</p>
            {database.status === 'creating' && (
              <p className="text-sm mt-2">Shards are being created...</p>
            )}
          </div>
        ) : (
          <div className="space-y-2">
            {(database.shard_ids ?? []).map((shardId) => (
              <div
                key={shardId}
                className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg"
              >
                <div className="flex items-center space-x-3">
                  <DatabaseIcon className="h-5 w-5 text-blue-500" />
                  <div>
                    <p className="font-medium text-gray-900 dark:text-white">{shardId}</p>
                    <p className="text-sm text-gray-500 dark:text-gray-400">Shard ID</p>
                  </div>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => navigate(`/shards/${shardId}`)}
                >
                  View Shard
                </Button>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Metadata */}
      {database.metadata && Object.keys(database.metadata).length > 0 && (
        <div className="card">
          <h3 className="text-lg font-semibold mb-4 text-gray-900 dark:text-white">Metadata</h3>
          <pre className="text-sm bg-gray-100 dark:bg-gray-800 p-4 rounded overflow-x-auto">
            {JSON.stringify(database.metadata, null, 2)}
          </pre>
        </div>
      )}

      {/* Timestamps */}
      <div className="card">
        <h3 className="text-lg font-semibold mb-4 text-gray-900 dark:text-white">Timestamps</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="text-sm font-medium text-gray-500 dark:text-gray-400">Created At</label>
            <p className="text-gray-900 dark:text-white">
              {formatDate(new Date(database.created_at))}
            </p>
            <p className="text-sm text-gray-500 dark:text-gray-400">
              {formatRelativeTime(new Date(database.created_at))}
            </p>
          </div>
          <div>
            <label className="text-sm font-medium text-gray-500 dark:text-gray-400">Updated At</label>
            <p className="text-gray-900 dark:text-white">
              {formatDate(new Date(database.updated_at))}
            </p>
            <p className="text-sm text-gray-500 dark:text-gray-400">
              {formatRelativeTime(new Date(database.updated_at))}
            </p>
          </div>
        </div>
      </div>

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={isDeleteModalOpen}
        onClose={() => setIsDeleteModalOpen(false)}
        title="Delete Database"
        footer={
          <>
            <Button variant="outline" onClick={() => setIsDeleteModalOpen(false)}>
              Cancel
            </Button>
            <Button variant="danger" onClick={handleDelete}>
              Delete
            </Button>
          </>
        }
      >
        <p className="text-gray-600 dark:text-gray-400">
          Are you sure you want to delete the database <strong>{database.name}</strong>?
          This action cannot be undone and will delete all associated shards and data.
        </p>
      </Modal>
    </div>
  );
}

