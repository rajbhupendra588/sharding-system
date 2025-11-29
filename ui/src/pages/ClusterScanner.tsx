import { useState } from 'react';
import { useClusters, useScanClusters, useScanResults, useCreateCluster, useDeleteCluster, useAvailableClusters, type AvailableCluster } from '@/features/cluster-scanner';
import { Plus, Search, Database, Trash2, Play, Cloud, HardDrive, RefreshCw, CheckCircle2 } from 'lucide-react';
import { toast } from 'react-hot-toast';
import LoadingSpinner from '@/components/ui/LoadingSpinner';
import Button from '@/components/ui/Button';
import Input from '@/components/ui/Input';
import Modal from '@/components/ui/Modal';
import StatusBadge from '@/components/ui/StatusBadge';
import { Table, TableHead, TableHeader, TableBody, TableRow, TableCell } from '@/components/ui/Table';
import { formatRelativeTime } from '@/lib/utils';
import type { CreateClusterRequest } from '@/features/cluster-scanner';

export default function ClusterScanner() {
  const [searchTerm, setSearchTerm] = useState('');
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedClusterIds, setSelectedClusterIds] = useState<string[]>([]);
  const [deepScan, setDeepScan] = useState(false);

  const { data: clusters, isLoading, error: clustersError } = useClusters();
  const { data: scanResults, error: scanResultsError } = useScanResults();
  const { data: availableClusters, isLoading: discovering, refetch: refetchAvailable } = useAvailableClusters();
  const scanMutation = useScanClusters();
  const createMutation = useCreateCluster();
  const deleteMutation = useDeleteCluster();

  const filteredClusters = clusters?.filter((cluster) =>
    cluster.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    cluster.id.toLowerCase().includes(searchTerm.toLowerCase())
  ) || [];

  const handleCreateCluster = async (data: CreateClusterRequest) => {
    try {
      await createMutation.mutateAsync(data);
      toast.success('Cluster registered successfully');
      setIsModalOpen(false);
    } catch (error: any) {
      toast.error(error.message || 'Failed to register cluster');
    }
  };

  const handleDeleteCluster = async (id: string) => {
    if (!confirm('Are you sure you want to delete this cluster?')) {
      return;
    }
    try {
      await deleteMutation.mutateAsync(id);
      toast.success('Cluster deleted successfully');
    } catch (error: any) {
      toast.error(error.message || 'Failed to delete cluster');
    }
  };

  const handleScan = async () => {
    try {
      const request = {
        cluster_ids: selectedClusterIds.length > 0 ? selectedClusterIds : undefined,
        deep_scan: deepScan,
      };
      await scanMutation.mutateAsync(request);
      toast.success('Scan started successfully');
      setSelectedClusterIds([]);
    } catch (error: any) {
      toast.error(error.message || 'Failed to start scan');
    }
  };

  const toggleClusterSelection = (id: string) => {
    setSelectedClusterIds((prev) =>
      prev.includes(id) ? prev.filter((c) => c !== id) : [...prev, id]
    );
  };

  if (isLoading) {
    return <LoadingSpinner size="lg" />;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Multi-Cluster Database Scanner</h1>
          <p className="text-gray-600 dark:text-gray-400 mt-1">
            Discover and scan databases across multiple Kubernetes clusters
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            onClick={() => refetchAvailable()}
            variant="secondary"
            className="flex items-center gap-2"
            disabled={discovering}
          >
            <RefreshCw className={`w-4 h-4 ${discovering ? 'animate-spin' : ''}`} />
            Discover Clusters
          </Button>
          <Button
            onClick={() => setIsModalOpen(true)}
            className="flex items-center gap-2"
          >
            <Plus className="w-4 h-4" />
            Register Cluster
          </Button>
          <Button
            onClick={handleScan}
            disabled={scanMutation.isPending || selectedClusterIds.length === 0}
            className="flex items-center gap-2"
            variant="primary"
          >
            <Play className="w-4 h-4" />
            {scanMutation.isPending ? 'Scanning...' : 'Start Scan'}
          </Button>
        </div>
      </div>

      {clustersError && (
        <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
          <p className="text-red-800 dark:text-red-200 font-medium">Error loading clusters</p>
          <p className="text-red-600 dark:text-red-300 text-sm mt-1">
            {clustersError instanceof Error ? clustersError.message : 'Failed to fetch clusters. Please check your API connection.'}
          </p>
        </div>
      )}

      {/* Available Clusters Section */}
      {availableClusters && availableClusters.length > 0 && (
        <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-6">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h2 className="text-xl font-semibold text-gray-900 dark:text-white flex items-center gap-2">
                <Cloud className="w-5 h-5" />
                Available Kubernetes Clusters
              </h2>
              <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                Clusters discovered from your kubeconfig file. Click "Register" to add them to the scanner.
              </p>
            </div>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {availableClusters
              .filter((cluster: AvailableCluster) => !cluster.is_registered)
              .map((cluster: AvailableCluster) => (
              <div
                key={cluster.context_name}
                className={`p-4 border rounded-lg ${
                  cluster.is_registered
                    ? 'border-green-200 bg-green-50 dark:border-green-800 dark:bg-green-900/20'
                    : 'border-gray-200 bg-white dark:border-gray-700 dark:bg-gray-800'
                }`}
              >
                <div className="flex items-start justify-between mb-2">
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <h3 className="font-medium text-gray-900 dark:text-white">{cluster.context_name}</h3>
                      {cluster.is_current && (
                        <span className="text-xs px-2 py-0.5 bg-blue-100 dark:bg-blue-900/30 rounded text-blue-700 dark:text-blue-300">
                          Current
                        </span>
                      )}
                      {cluster.is_registered && (
                        <span className="flex items-center gap-1 text-xs text-green-600 dark:text-green-400">
                          <CheckCircle2 className="w-3 h-3" />
                          Registered
                        </span>
                      )}
                    </div>
                    <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                      Cluster: {cluster.cluster_name}
                    </p>
                    <p className="text-xs text-gray-500 dark:text-gray-400">
                      Type: {cluster.type} • Provider: {cluster.provider}
                    </p>
                  </div>
                </div>
                {!cluster.is_registered && (
                  <Button
                    size="sm"
                    onClick={() => {
                      handleCreateCluster({
                        name: cluster.context_name,
                        type: cluster.type as 'cloud' | 'onprem',
                        provider: cluster.provider,
                        context: cluster.context_name,
                        kubeconfig: '',
                        endpoint: '',
                      });
                    }}
                    className="w-full mt-2"
                    disabled={createMutation.isPending}
                  >
                    <Plus className="w-3 h-3 mr-1" />
                    Register
                  </Button>
                )}
              </div>
            ))}
          </div>
        </div>
      )}

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-4">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
              <Input
                type="text"
                placeholder="Search clusters..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-10 w-64"
              />
            </div>
            {selectedClusterIds.length > 0 && (
              <span className="text-sm text-gray-600 dark:text-gray-400">
                {selectedClusterIds.length} cluster(s) selected
              </span>
            )}
          </div>
          <label className="flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              checked={deepScan}
              onChange={(e) => setDeepScan(e.target.checked)}
              className="rounded"
            />
            <span className="text-sm">Deep Scan</span>
          </label>
        </div>

        <Table>
          <TableHead>
            <TableHeader className="w-12">
              <input
                type="checkbox"
                checked={selectedClusterIds.length === filteredClusters.length && filteredClusters.length > 0}
                onChange={(e) => {
                  if (e.target.checked) {
                    setSelectedClusterIds(filteredClusters.map((c) => c.id));
                  } else {
                    setSelectedClusterIds([]);
                  }
                }}
              />
            </TableHeader>
            <TableHeader>Name</TableHeader>
            <TableHeader>Type</TableHeader>
            <TableHeader>Provider</TableHeader>
            <TableHeader>Status</TableHeader>
            <TableHeader>Last Scan</TableHeader>
            <TableHeader>Actions</TableHeader>
          </TableHead>
          <TableBody>
            {filteredClusters.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="text-center py-12">
                  <div className="flex flex-col items-center justify-center space-y-4">
                    <div className="bg-blue-100 dark:bg-blue-900/30 w-16 h-16 rounded-full flex items-center justify-center">
                      <Cloud className="h-8 w-8 text-blue-600 dark:text-blue-400" />
                    </div>
                    <div>
                      <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
                        No Kubernetes Clusters Registered
                      </h3>
                      <p className="text-sm text-gray-600 dark:text-gray-400 max-w-md mb-4">
                        Register a Kubernetes cluster to start scanning for databases. You can register local clusters (minikube, kind, Docker Desktop) or cloud clusters (EKS, GKE, AKS).
                      </p>
                      <div className="text-left bg-gray-50 dark:bg-gray-900/50 rounded-lg p-4 max-w-md mx-auto mb-4">
                        <p className="text-xs font-medium text-gray-700 dark:text-gray-300 mb-2">Quick Start:</p>
                        <ol className="text-xs text-gray-600 dark:text-gray-400 space-y-1 list-decimal list-inside">
                          <li>Click "Register Cluster" button above</li>
                          <li>Enter a cluster name (e.g., "local-cluster")</li>
                          <li>Select cluster type (On-Premises or Cloud)</li>
                          <li>For local clusters, you can leave kubeconfig empty if using default context</li>
                          <li>For remote clusters, provide kubeconfig path or content</li>
                        </ol>
                      </div>
                      <Button
                        onClick={() => setIsModalOpen(true)}
                        className="flex items-center gap-2"
                        variant="primary"
                      >
                        <Plus className="w-4 h-4" />
                        Register Your First Cluster
                      </Button>
                    </div>
                  </div>
                </TableCell>
              </TableRow>
            ) : (
              filteredClusters.map((cluster) => (
                <TableRow key={cluster.id}>
                  <TableCell>
                    <input
                      type="checkbox"
                      checked={selectedClusterIds.includes(cluster.id)}
                      onChange={() => toggleClusterSelection(cluster.id)}
                    />
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      {cluster.type === 'cloud' ? (
                        <Cloud className="w-4 h-4 text-blue-500" />
                      ) : (
                        <HardDrive className="w-4 h-4 text-gray-500" />
                      )}
                      <span className="font-medium">{cluster.name}</span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <span className="text-sm capitalize">{cluster.type}</span>
                  </TableCell>
                  <TableCell>
                    <span className="text-sm">{cluster.provider || 'N/A'}</span>
                  </TableCell>
                  <TableCell>
                    <StatusBadge status={cluster.status} />
                  </TableCell>
                  <TableCell>
                    {cluster.last_scan ? (
                      <span className="text-sm text-gray-600 dark:text-gray-400">
                        {formatRelativeTime(new Date(cluster.last_scan))}
                      </span>
                    ) : (
                      <span className="text-sm text-gray-400">Never</span>
                    )}
                  </TableCell>
                  <TableCell>
                    <Button
                      onClick={() => handleDeleteCluster(cluster.id)}
                      variant="secondary"
                      size="sm"
                      className="text-red-600 hover:text-red-700"
                    >
                      <Trash2 className="w-4 h-4" />
                    </Button>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {scanResultsError && (
        <div className="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg p-4">
          <p className="text-yellow-800 dark:text-yellow-200 font-medium">Warning: Could not load scan results</p>
          <p className="text-yellow-600 dark:text-yellow-300 text-sm mt-1">
            {scanResultsError instanceof Error ? scanResultsError.message : 'Failed to fetch scan results.'}
          </p>
        </div>
      )}

      {scanResults && scanResults.length > 0 && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold mb-4">Scan Results</h2>
          <Table>
            <TableHead>
              <TableHeader>Database</TableHeader>
              <TableHeader>Cluster</TableHeader>
              <TableHeader>Namespace</TableHeader>
              <TableHeader>Host</TableHeader>
              <TableHeader>Status</TableHeader>
              <TableHeader>Discovered</TableHeader>
            </TableHead>
            <TableBody>
              {scanResults.map((db) => (
                <TableRow key={db.id}>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <Database className="w-4 h-4" />
                      <span className="font-medium">{db.database_name}</span>
                    </div>
                  </TableCell>
                  <TableCell>{db.cluster_name}</TableCell>
                  <TableCell>{db.namespace}</TableCell>
                  <TableCell>{db.host}:{db.port}</TableCell>
                  <TableCell>
                    <StatusBadge status={db.status} />
                  </TableCell>
                  <TableCell>
                    {formatRelativeTime(new Date(db.discovered_at))}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      <RegisterClusterModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSubmit={handleCreateCluster}
      />
    </div>
  );
}

function RegisterClusterModal({
  isOpen,
  onClose,
  onSubmit,
}: {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: CreateClusterRequest) => void;
}) {
  const [formData, setFormData] = useState<CreateClusterRequest>({
    name: '',
    type: 'onprem',
    provider: '',
    kubeconfig: '',
    context: '',
    endpoint: '',
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(formData);
    setFormData({
      name: '',
      type: 'onprem',
      provider: '',
      kubeconfig: '',
      context: '',
      endpoint: '',
    });
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Register Kubernetes Cluster">
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium mb-1">Cluster Name *</label>
          <Input
            required
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            placeholder="my-cluster"
          />
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">Type *</label>
          <select
            required
            value={formData.type}
            onChange={(e) => setFormData({ ...formData, type: e.target.value as 'cloud' | 'onprem' })}
            className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600"
          >
            <option value="onprem">On-Premises</option>
            <option value="cloud">Cloud</option>
          </select>
        </div>

        {formData.type === 'cloud' && (
          <div>
            <label className="block text-sm font-medium mb-1">Provider</label>
            <select
              value={formData.provider || ''}
              onChange={(e) => setFormData({ ...formData, provider: e.target.value })}
              className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600"
            >
              <option value="">Select provider</option>
              <option value="aws">AWS (EKS)</option>
              <option value="gcp">GCP (GKE)</option>
              <option value="azure">Azure (AKS)</option>
            </select>
          </div>
        )}

        <div>
          <label className="block text-sm font-medium mb-1">
            Kubeconfig Path or Content
            <span className="text-xs text-gray-500 dark:text-gray-400 ml-2">(Optional for local clusters)</span>
          </label>
          <textarea
            value={formData.kubeconfig || ''}
            onChange={(e) => setFormData({ ...formData, kubeconfig: e.target.value })}
            placeholder="/path/to/kubeconfig or paste kubeconfig content (leave empty for default local context)"
            className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600"
            rows={4}
          />
          <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
            For local clusters (minikube, kind, Docker Desktop), you can leave this empty to use your default kubeconfig context.
          </p>
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">Context Name</label>
          <Input
            value={formData.context || ''}
            onChange={(e) => setFormData({ ...formData, context: e.target.value })}
            placeholder="my-context"
          />
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">API Endpoint</label>
          <Input
            value={formData.endpoint || ''}
            onChange={(e) => setFormData({ ...formData, endpoint: e.target.value })}
            placeholder="https://api.cluster.example.com"
          />
        </div>

        <div className="flex justify-end gap-2 pt-4">
          <Button type="button" onClick={onClose} variant="secondary">
            Cancel
          </Button>
          <Button type="submit" variant="primary">
            Register Cluster
          </Button>
        </div>
      </form>
    </Modal>
  );
}

