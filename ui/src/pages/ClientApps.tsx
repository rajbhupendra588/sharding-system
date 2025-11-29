import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Users, Plus, Search, Database, Activity, Clock, Server, RefreshCw, CheckCircle2, AlertCircle, Trash2, CheckSquare, Square } from 'lucide-react';
import { apiClient } from '@/lib/api-client';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import Modal from '@/components/ui/Modal';
import { useQuery } from '@tanstack/react-query';
import { useClientApps } from '@/features/clientApp';
import { useShards } from '@/features/shard';
import LoadingSpinner from '@/components/ui/LoadingSpinner';
import StatusBadge from '@/components/ui/StatusBadge';
import Button from '@/components/ui/Button';
import { formatRelativeTime, formatDate } from '@/lib/utils';
import { toast } from 'react-hot-toast';
import ConnectAppModal from '@/components/client-app/ConnectAppModal';

interface DiscoveredApp {
  namespace: string;
  name: string;
  type: string;
  database_name: string;
  database_url?: string;
  database_host?: string;
  database_port?: string;
  database_user?: string;
  labels: Record<string, string>;
  annotations: Record<string, string>;
  is_registered: boolean;
}

export default function ClientApps() {
  const { data: clientApps, isLoading, error, refetch } = useClientApps();
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showDiscoveredApps, setShowDiscoveredApps] = useState(true);
  const [newApp, setNewApp] = useState({ name: '', description: '', database_name: '', key_prefix: '' });
  const [deleteAppId, setDeleteAppId] = useState<string | null>(null);
  const [selectedApps, setSelectedApps] = useState<Set<string>>(new Set());
  const [showBulkConfirm, setShowBulkConfirm] = useState(false);
  const [connectApp, setConnectApp] = useState<{ name: string; database: string } | null>(null);

  const deleteMutation = useMutation({
    mutationFn: (id: string) => apiClient.deleteClientApp(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['clientApps'] });
      setDeleteAppId(null);
      toast.success('Client application de-registered successfully');
      refetch();
    },
    onError: (error: { message: string }) => {
      toast.error(`Failed to de-register application: ${error.message}`);
    },
  });

  const bulkDeleteMutation = useMutation({
    mutationFn: async (ids: string[]) => {
      await Promise.all(ids.map(id => apiClient.deleteClientApp(id)));
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['clientApps'] });
      setSelectedApps(new Set());
      setShowBulkConfirm(false);
      toast.success('Client applications de-registered successfully');
      refetch();
    },
    onError: (error: { message: string }) => {
      toast.error(`Failed to de-register applications: ${error.message}`);
    },
  });

  // Fetch discovered applications from Kubernetes
  const { data: discoveredApps, isLoading: discovering, refetch: refetchDiscovery } = useQuery<DiscoveredApp[]>({
    queryKey: ['discoveredApps'],
    queryFn: async () => {
      try {
        const response = await fetch('/api/v1/client-apps/discover');
        if (!response.ok) {
          if (response.status === 503) {
            // Kubernetes discovery not available - return empty array
            return [];
          }
          if (response.status === 404) {
            // Endpoint not found - manager might not be running or route not registered
            console.warn('Discovery endpoint not found. Is the manager running?');
            return [];
          }
          throw new Error(`Failed to discover applications: ${response.status} ${response.statusText}`);
        }
        const data = await response.json();
        // Ensure we always return an array, handle null/undefined gracefully
        return Array.isArray(data) ? data : [];
      } catch (error: any) {
        // Network errors or other issues - return empty array gracefully
        console.warn('Discovery failed:', error.message);
        return [];
      }
    },
    retry: false,
    refetchOnWindowFocus: false,
  });

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <LoadingSpinner size="lg" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="card">
        <div className="text-center py-12">
          <p className="text-red-600 dark:text-red-400">Failed to load client applications</p>
        </div>
      </div>
    );
  }

  const filteredApps = clientApps?.filter((app) =>
    app.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    app.description?.toLowerCase().includes(searchQuery.toLowerCase())
  ) || [];

  const handleCreate = async () => {
    try {
      const response = await fetch('/api/v1/client-apps', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(newApp),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ message: 'Failed to create client application' }));
        throw new Error(errorData.message || 'Failed to create client application');
      }

      setShowCreateModal(false);
      setNewApp({ name: '', description: '', database_name: '', key_prefix: '' });
      toast.success('Client application registered successfully');
      refetch();
      refetchDiscovery();
    } catch (err: any) {
      toast.error(err.message || 'Failed to create client application');
    }
  };

  const handleAutoRegister = async (discoveredApp: DiscoveredApp) => {
    try {
      const appData = {
        name: discoveredApp.name,
        description: `Auto-registered from Kubernetes namespace: ${discoveredApp.namespace}`,
        database_name: discoveredApp.database_name || `${discoveredApp.name}_db`,
        database_host: discoveredApp.database_host || '',
        database_port: discoveredApp.database_port || '',
        database_user: discoveredApp.database_user || '',
        database_password: '', // Password not available from discovery
        key_prefix: discoveredApp.annotations['sharding.keyPrefix'] || '',
        namespace: discoveredApp.namespace,
      };

      const response = await fetch('/api/v1/client-apps', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(appData),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ message: 'Failed to register application' }));
        throw new Error(errorData.message || 'Failed to register application');
      }

      toast.success(`Successfully registered ${discoveredApp.name}`);
      refetch();
      refetchDiscovery();
    } catch (err: any) {
      toast.error(err.message || 'Failed to register application');
    }
  };

  const toggleAppSelection = (appId: string) => {
    const newSelected = new Set(selectedApps);
    if (newSelected.has(appId)) {
      newSelected.delete(appId);
    } else {
      newSelected.add(appId);
    }
    setSelectedApps(newSelected);
  };

  const toggleSelectAll = () => {
    if (selectedApps.size === filteredApps.length) {
      setSelectedApps(new Set());
    } else {
      setSelectedApps(new Set(filteredApps.map(a => a.id)));
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white">Client Applications</h1>
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {selectedApps.size > 0
              ? `${selectedApps.size} application${selectedApps.size !== 1 ? 's' : ''} selected`
              : 'Manage client applications using the sharding system'
            }
          </p>
        </div>
        <div className="flex items-center gap-2">
          {selectedApps.size > 0 && (
            <Button
              variant="danger"
              onClick={() => setShowBulkConfirm(true)}
              disabled={bulkDeleteMutation.isPending}
            >
              <Trash2 className="h-4 w-4 mr-2" />
              De-register ({selectedApps.size})
            </Button>
          )}
          <Button
            variant="secondary"
            onClick={() => {
              refetchDiscovery();
              setShowDiscoveredApps(true);
            }}
            disabled={discovering}
          >
            <RefreshCw className={`h-4 w-4 mr-2 ${discovering ? 'animate-spin' : ''}`} />
            Discover Apps
          </Button>
          <Button onClick={() => setShowCreateModal(true)}>
            <Plus className="h-4 w-4 mr-2" />
            Register Manually
          </Button>
        </div>
      </div>

      {/* Discovered Applications Section */}
      {showDiscoveredApps && discoveredApps && discoveredApps.length > 0 && (
        <div className="card">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h2 className="text-xl font-semibold text-gray-900 dark:text-white flex items-center gap-2">
                <Database className="h-5 w-5" />
                Discovered Applications ({discoveredApps.length})
              </h2>
              <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                Applications and databases automatically discovered from Kubernetes namespaces
              </p>
            </div>
            <button
              onClick={() => setShowDiscoveredApps(false)}
              className="text-sm text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
            >
              Hide
            </button>
          </div>

          <div className="space-y-3">
            {discoveredApps.map((app) => (
              <div
                key={`${app.namespace}-${app.name}`}
                className={`p-4 border rounded-lg ${app.is_registered
                  ? 'border-green-200 bg-green-50 dark:border-green-800 dark:bg-green-900/20'
                  : 'border-gray-200 bg-gray-50 dark:border-gray-700 dark:bg-gray-800'
                  }`}
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-2">
                      <h3 className="font-medium text-gray-900 dark:text-white">{app.name}</h3>
                      <span className="text-xs px-2 py-1 bg-gray-200 dark:bg-gray-700 rounded text-gray-600 dark:text-gray-300">
                        {app.namespace}
                      </span>
                      <span className="text-xs px-2 py-1 bg-blue-100 dark:bg-blue-900/30 rounded text-blue-700 dark:text-blue-300">
                        {app.type}
                      </span>
                      {app.is_registered && (
                        <span className="flex items-center gap-1 text-xs text-green-600 dark:text-green-400">
                          <CheckCircle2 className="h-3 w-3" />
                          Registered
                        </span>
                      )}
                    </div>
                    <div className="space-y-1 text-sm">
                      <div className="flex items-center gap-2">
                        <Database className="h-4 w-4 text-gray-400" />
                        <span className="text-gray-600 dark:text-gray-400">Database:</span>
                        <span className="font-medium text-gray-900 dark:text-white">
                          {app.database_name || 'Not detected'}
                        </span>
                      </div>
                      {app.database_host && (
                        <div className="flex items-center gap-2 text-gray-600 dark:text-gray-400">
                          <Server className="h-4 w-4" />
                          <span>
                            {app.database_host}
                            {app.database_port && `:${app.database_port}`}
                          </span>
                        </div>
                      )}
                    </div>
                  </div>
                  {!app.is_registered && (
                    <Button
                      size="sm"
                      onClick={() => handleAutoRegister(app)}
                      className="ml-4"
                    >
                      <Plus className="h-4 w-4 mr-1" />
                      Register
                    </Button>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {showDiscoveredApps && discoveredApps && discoveredApps.length === 0 && !discovering && (
        <div className="card">
          <div className="text-center py-8">
            <AlertCircle className="h-12 w-12 text-gray-400 mx-auto mb-3" />
            <p className="text-gray-600 dark:text-gray-400">
              No applications discovered. Make sure you're running in a Kubernetes cluster with proper permissions.
            </p>
          </div>
        </div>
      )}

      {/* Search */}
      <div className="card">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-gray-400" />
          <input
            type="text"
            placeholder="Search client applications..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary-500 focus:border-transparent"
          />
        </div>
      </div>

      {/* Client Apps List */}
      {filteredApps.length > 0 ? (
        <>
          <div className="flex items-center justify-between mb-4">
            <button
              onClick={toggleSelectAll}
              className="flex items-center gap-2 px-3 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
            >
              {selectedApps.size === filteredApps.length && filteredApps.length > 0 ? (
                <CheckSquare className="h-5 w-5 text-primary-600" />
              ) : (
                <Square className="h-5 w-5 text-gray-400" />
              )}
              {selectedApps.size === filteredApps.length && filteredApps.length > 0 ? 'Deselect All' : 'Select All'}
            </button>
          </div>
          <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
            {filteredApps.map((app) => (
              <div key={app.id} className="card hover:shadow-lg transition-shadow relative">
                <div className="absolute top-4 right-4">
                  <button
                    onClick={() => toggleAppSelection(app.id)}
                    className="p-1 hover:bg-gray-100 dark:hover:bg-gray-700 rounded"
                  >
                    {selectedApps.has(app.id) ? (
                      <CheckSquare className="h-5 w-5 text-primary-600" />
                    ) : (
                      <Square className="h-5 w-5 text-gray-400" />
                    )}
                  </button>
                </div>
                <div className="flex items-start justify-between mb-4 pr-10">
                  <div className="flex items-center gap-3">
                    <div className="p-2 bg-primary-50 dark:bg-primary-900/20 rounded-lg">
                      <Users className="h-5 w-5 text-primary-600 dark:text-primary-400" />
                    </div>
                    <div>
                      <h3 className="text-lg font-semibold text-gray-900 dark:text-white">{app.name}</h3>
                      {app.description && (
                        <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">{app.description}</p>
                      )}
                      <div className="flex gap-2 mt-1">
                        {app.namespace && (
                          <span className="text-xs px-2 py-0.5 bg-gray-100 dark:bg-gray-700 rounded text-gray-600 dark:text-gray-300 border border-gray-200 dark:border-gray-600">
                            ns: {app.namespace}
                          </span>
                        )}
                        {app.cluster_name && (
                          <span className="text-xs px-2 py-0.5 bg-blue-50 dark:bg-blue-900/20 rounded text-blue-600 dark:text-blue-300 border border-blue-100 dark:border-blue-800">
                            cluster: {app.cluster_name}
                          </span>
                        )}
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <StatusBadge status={app.status === 'active' ? 'healthy' : 'inactive'} />
                    <div className="flex items-center gap-2">
                      <Button
                        variant="secondary"
                        size="sm"
                        onClick={() => setConnectApp({ name: app.name, database: app.database_name || '' })}
                      >
                        <Database className="h-4 w-4 mr-2" />
                        Connect
                      </Button>
                      <Button
                        variant="danger"
                        size="sm"
                        onClick={() => setDeleteAppId(app.id)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                </div>
                {/* The original second delete button is removed as it's now part of the actions above */}

                <div className="space-y-3">
                  {app.database_name && (
                    <div className="flex items-center gap-2 text-sm">
                      <Database className="h-4 w-4 text-gray-400" />
                      <span className="text-gray-500 dark:text-gray-400">Database:</span>
                      <span className="font-medium text-gray-900 dark:text-white">{app.database_name}</span>
                    </div>
                  )}
                  {app.key_prefix && (
                    <div className="flex items-center gap-2 text-sm">
                      <span className="text-gray-500 dark:text-gray-400">Key Prefix:</span>
                      <code className="px-2 py-1 bg-gray-100 dark:bg-gray-700 rounded text-gray-900 dark:text-gray-100">
                        {app.key_prefix}
                      </code>
                    </div>
                  )}

                  <div className="grid grid-cols-2 gap-4 pt-3 border-t border-gray-200 dark:border-gray-700">
                    <div>
                      <div className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400 mb-1">
                        <Database className="h-4 w-4" />
                        Shards Used
                      </div>
                      <p className="text-lg font-semibold text-gray-900 dark:text-white">
                        {app.shard_ids?.length || 0}
                      </p>
                    </div>
                    <div>
                      <div className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400 mb-1">
                        <Activity className="h-4 w-4" />
                        Requests
                      </div>
                      <p className="text-lg font-semibold text-gray-900 dark:text-white">
                        {app.request_count?.toLocaleString() || 0}
                      </p>
                    </div>
                  </div>

                  <div className="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400 pt-2 border-t border-gray-200 dark:border-gray-700">
                    <div className="flex items-center gap-1">
                      <Clock className="h-3 w-3" />
                      Last seen: {formatRelativeTime(app.last_seen)}
                    </div>
                    <span>Created: {formatDate(app.created_at)}</span>
                  </div>

                  {app.shard_ids && app.shard_ids.length > 0 && (
                    <div className="pt-3 border-t border-gray-200 dark:border-gray-700">
                      <div className="flex items-center justify-between mb-2">
                        <p className="text-xs font-medium text-gray-500 dark:text-gray-400">Shards & Databases:</p>
                        <Link
                          to={`/shards?filter=${app.id}`}
                          className="text-xs text-primary-600 hover:text-primary-700"
                        >
                          View all →
                        </Link>
                      </div>
                      <ShardListForClient clientAppId={app.id} shardIds={app.shard_ids} />
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        </>
      ) : (
        <div className="card border-dashed border-2 border-gray-300 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50">
          <div className="text-center py-12 px-4">
            <div className="bg-blue-100 dark:bg-blue-900/30 w-16 h-16 rounded-full flex items-center justify-center mx-auto mb-4">
              <Server className="h-8 w-8 text-blue-600 dark:text-blue-400" />
            </div>
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">No Registered Applications</h3>
            <p className="mt-2 text-sm text-gray-600 dark:text-gray-300 max-w-md mx-auto">
              {searchQuery
                ? 'No applications match your search.'
                : 'De-registering an app only removes it from this list. Your actual applications and databases are still running safely in your Kubernetes cluster.'}
            </p>

            {!searchQuery && (
              <div className="mt-8 flex flex-col sm:flex-row gap-4 justify-center">
                <Button
                  onClick={() => {
                    refetchDiscovery();
                    setShowDiscoveredApps(true);
                  }}
                  disabled={discovering}
                  className="flex items-center"
                >
                  <RefreshCw className={`h-4 w-4 mr-2 ${discovering ? 'animate-spin' : ''}`} />
                  Scan Cluster for Apps
                </Button>
                <Button
                  variant="secondary"
                  onClick={() => setShowCreateModal(true)}
                >
                  <Plus className="h-4 w-4 mr-2" />
                  Manually Register
                </Button>
              </div>
            )}

            {!searchQuery && (
              <p className="mt-6 text-xs text-gray-500 dark:text-gray-400">
                Tip: Click "Scan Cluster" to find and re-register your existing applications.
              </p>
            )}
          </div>
        </div>
      )}

      {/* Create Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white dark:bg-gray-800 rounded-lg p-6 w-full max-w-md">
            <h2 className="text-xl font-bold text-gray-900 dark:text-white mb-4">Register Client Application</h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Name *
                </label>
                <input
                  type="text"
                  value={newApp.name}
                  onChange={(e) => setNewApp({ ...newApp, name: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                  placeholder="e.g., E-commerce Service"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Description
                </label>
                <textarea
                  value={newApp.description}
                  onChange={(e) => setNewApp({ ...newApp, description: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                  rows={3}
                  placeholder="Optional description"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Database Name <span className="text-red-500">*</span>
                </label>
                <input
                  type="text"
                  value={newApp.database_name}
                  onChange={(e) => setNewApp({ ...newApp, database_name: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                  placeholder="e.g., ecommerce_db, analytics_db"
                  required
                />
                <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                  Database name for which sharding needs to be created
                </p>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Key Prefix (Optional)
                </label>
                <input
                  type="text"
                  value={newApp.key_prefix}
                  onChange={(e) => setNewApp({ ...newApp, key_prefix: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                  placeholder="e.g., app1:"
                />
                <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                  Shard keys starting with this prefix will be associated with this client
                </p>
              </div>
            </div>
            <div className="flex gap-3 mt-6">
              <button
                onClick={() => {
                  setShowCreateModal(false);
                  setNewApp({ name: '', description: '', database_name: '', key_prefix: '' });
                }}
                className="flex-1 px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700"
              >
                Cancel
              </button>
              <button
                onClick={handleCreate}
                disabled={!newApp.name || !newApp.database_name}
                className="flex-1 px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Register
              </button>
            </div>
          </div>
        </div>
      )}

      {/* De-register Confirmation Modal */}
      <Modal
        isOpen={deleteAppId !== null}
        onClose={() => setDeleteAppId(null)}
        title="De-register Application"
        size="sm"
        footer={
          <>
            <Button
              variant="secondary"
              onClick={() => setDeleteAppId(null)}
            >
              Cancel
            </Button>
            <Button
              variant="danger"
              onClick={() => {
                if (deleteAppId) {
                  deleteMutation.mutate(deleteAppId);
                }
              }}
              isLoading={deleteMutation.isPending}
            >
              De-register
            </Button>
          </>
        }
      >
        <p className="text-sm text-gray-600 dark:text-gray-300">
          Are you sure you want to de-register this client application?
          <br /><br />
          This will remove the application from the Sharding Manager registry. It will <strong>not</strong> delete the actual application or its database.
        </p>
      </Modal>

      {/* Bulk De-register Confirmation Modal */}
      <Modal
        isOpen={showBulkConfirm}
        onClose={() => setShowBulkConfirm(false)}
        title="De-register Applications"
        size="sm"
        footer={
          <>
            <Button
              variant="secondary"
              onClick={() => setShowBulkConfirm(false)}
            >
              Cancel
            </Button>
            <Button
              variant="danger"
              onClick={() => bulkDeleteMutation.mutate(Array.from(selectedApps))}
              isLoading={bulkDeleteMutation.isPending}
            >
              De-register {selectedApps.size} Application{selectedApps.size !== 1 ? 's' : ''}
            </Button>
          </>
        }
      >
        <p className="text-sm text-gray-600 dark:text-gray-300">
          Are you sure you want to de-register {selectedApps.size} client application{selectedApps.size !== 1 ? 's' : ''}?
          <br /><br />
          This will remove the application{selectedApps.size !== 1 ? 's' : ''} from the Sharding Manager registry. It will <strong>not</strong> delete the actual application{selectedApps.size !== 1 ? 's' : ''} or {selectedApps.size !== 1 ? 'their' : 'its'} database.
        </p>
      </Modal>

      {/* Connect App Modal */}
      {connectApp && (
        <ConnectAppModal
          isOpen={true}
          onClose={() => setConnectApp(null)}
          appName={connectApp.name}
          databaseName={connectApp.database}
        />
      )}
    </div>
  );
}

// Component to show detailed shard information for a client app
function ShardListForClient({ clientAppId, shardIds }: { clientAppId: string; shardIds: string[] }) {
  const { data: allShards } = useShards();

  // Filter shards that belong to this client app
  const clientShards = allShards?.filter(shard =>
    shard.client_app_id === clientAppId && shardIds.includes(shard.id)
  ) || [];

  if (clientShards.length === 0) {
    return (
      <div className="text-xs text-gray-500 dark:text-gray-400">
        Loading shard details...
      </div>
    );
  }

  return (
    <div className="space-y-2">
      {clientShards.slice(0, 3).map((shard) => (
        <Link
          key={shard.id}
          to={`/shards/${shard.id}`}
          className="block p-2 bg-gray-50 dark:bg-gray-800 rounded hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
        >
          <div className="flex items-start justify-between">
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2 mb-1">
                <Database className="h-3 w-3 text-primary-600 dark:text-primary-400 flex-shrink-0" />
                <span className="text-xs font-medium text-gray-900 dark:text-white truncate">
                  {shard.name}
                </span>
                <StatusBadge status={shard.status} />
              </div>
              <div className="flex items-center gap-2 text-xs text-gray-600 dark:text-gray-400">
                <Server className="h-3 w-3" />
                <span className="truncate" title={shard.primary_endpoint}>
                  {shard.primary_endpoint.replace(/^postgres(ql)?:\/\/([^:]+):([^@]+)@/, 'postgres://***:***@')}
                </span>
              </div>
            </div>
          </div>
        </Link>
      ))}
      {clientShards.length > 3 && (
        <Link
          to={`/shards?filter=${clientAppId}`}
          className="block text-xs text-primary-600 hover:text-primary-700 text-center py-1"
        >
          +{clientShards.length - 3} more shards
        </Link>
      )}
    </div>
  );
}
