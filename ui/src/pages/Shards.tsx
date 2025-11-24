import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate, Link } from 'react-router-dom';
import { Plus, Search, Trash2, ExternalLink, Users, Database } from 'lucide-react';
import { apiClient } from '@/lib/api-client';
import { toast } from 'react-hot-toast';
import LoadingSpinner from '@/components/ui/LoadingSpinner';
import Button from '@/components/ui/Button';
import Input from '@/components/ui/Input';
import Modal from '@/components/ui/Modal';
import StatusBadge from '@/components/ui/StatusBadge';
import { Table, TableHead, TableHeader, TableBody, TableRow, TableCell } from '@/components/ui/Table';
import { formatRelativeTime } from '@/lib/utils';
import { useClientApps } from '@/features/clientApp';
import type { CreateShardRequest } from '@/types';

export default function Shards() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedClientApp, _setSelectedClientApp] = useState<string>(''); // Filter by client app
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [deleteShardId, setDeleteShardId] = useState<string | null>(null);

  const { data: shards, isLoading } = useQuery({
    queryKey: ['shards'],
    queryFn: () => apiClient.listShards(),
    refetchInterval: 10000,
  });

  const { data: clientApps } = useClientApps();

  const createMutation = useMutation({
    mutationFn: (data: CreateShardRequest) => apiClient.createShard(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['shards'] });
      setIsCreateModalOpen(false);
      toast.success('Shard created successfully');
    },
    onError: (error: { message: string }) => {
      toast.error(`Failed to create shard: ${error.message}`);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => apiClient.deleteShard(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['shards'] });
      setDeleteShardId(null);
      toast.success('Shard deleted successfully');
    },
    onError: (error: { message: string }) => {
      toast.error(`Failed to delete shard: ${error.message}`);
    },
  });

  // Get client app name mapping
  const clientAppMap = new Map(clientApps?.map(app => [app.id, app.name]) || []);

  const filteredShards = shards?.filter((shard) => {
    const matchesSearch = shard.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      shard.id.toLowerCase().includes(searchTerm.toLowerCase()) ||
      (shard.client_app_id && clientAppMap.get(shard.client_app_id)?.toLowerCase().includes(searchTerm.toLowerCase()));

    const matchesClientApp = !selectedClientApp || shard.client_app_id === selectedClientApp;

    return matchesSearch && matchesClientApp;
  }) || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">Shards</h1>
        <Button onClick={() => setIsCreateModalOpen(true)}>
          <Plus className="h-4 w-4 mr-2" />
          Create Shard
        </Button>
      </div>

      {/* Search */}
      <div className="card">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-gray-400" />
          <Input
            type="text"
            placeholder="Search shards by name or ID..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-10"
          />
        </div>
      </div>

      {/* Shards Table */}
      <div className="card">
        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <LoadingSpinner size="lg" />
          </div>
        ) : filteredShards.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-gray-500 dark:text-gray-400">No shards found</p>
          </div>
        ) : (
          <Table>
            <TableHead>
              <TableHeader>Name</TableHeader>
              <TableHeader>Client Application</TableHeader>
              <TableHeader>Database</TableHeader>
              <TableHeader>Status</TableHeader>
              <TableHeader>Replicas</TableHeader>
              <TableHeader>Created</TableHeader>
              <TableHeader>Actions</TableHeader>
            </TableHead>
            <TableBody>
              {filteredShards.map((shard) => {
                const clientAppName = shard.client_app_id ? clientAppMap.get(shard.client_app_id) : 'Unknown';
                return (
                  <TableRow
                    key={shard.id}
                    onClick={() => navigate(`/shards/${shard.id}`)}
                    className="hover:bg-gray-50 dark:hover:bg-gray-700/50"
                  >
                    <TableCell>
                      <div>
                        <div className="font-medium text-gray-900 dark:text-white">{shard.name}</div>
                        <div className="text-xs text-gray-500 dark:text-gray-400 font-mono">{shard.id.substring(0, 8)}...</div>
                      </div>
                    </TableCell>
                    <TableCell>
                      {shard.client_app_id ? (
                        <Link
                          to={`/client-apps?filter=${shard.client_app_id}`}
                          onClick={(e) => e.stopPropagation()}
                          className="flex items-center gap-2 text-sm text-primary-600 hover:text-primary-700 dark:text-primary-400"
                        >
                          <Users className="h-4 w-4" />
                          <span>{clientAppName || shard.client_app_id.substring(0, 8)}</span>
                        </Link>
                      ) : (
                        <span className="text-sm text-gray-400">No client app</span>
                      )}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Database className="h-4 w-4 text-gray-400" />
                        <div className="text-sm text-gray-900 dark:text-gray-300 max-w-xs truncate" title={shard.primary_endpoint}>
                          {shard.primary_endpoint.replace(/^postgres(ql)?:\/\/([^:]+):([^@]+)@/, 'postgres://***:***@')}
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <StatusBadge status={shard.status} />
                    </TableCell>
                    <TableCell>
                      <span className="text-sm text-gray-900 dark:text-gray-300">{shard.replicas?.length || 0}</span>
                    </TableCell>
                    <TableCell>
                      <div className="text-sm text-gray-500 dark:text-gray-400">
                        {formatRelativeTime(shard.created_at)}
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center space-x-2" onClick={(e) => e.stopPropagation()}>
                        <button
                          onClick={() => navigate(`/shards/${shard.id}`)}
                          className="text-primary-600 hover:text-primary-700"
                          title="View details"
                        >
                          <ExternalLink className="h-4 w-4" />
                        </button>
                        <button
                          onClick={() => setDeleteShardId(shard.id)}
                          className="text-red-600 hover:text-red-700"
                          title="Delete shard"
                        >
                          <Trash2 className="h-4 w-4" />
                        </button>
                      </div>
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        )}
      </div>

      {/* Create Shard Modal */}
      <CreateShardModal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        onSubmit={(data) => createMutation.mutate(data)}
        isLoading={createMutation.isPending}
      />

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={deleteShardId !== null}
        onClose={() => setDeleteShardId(null)}
        title="Delete Shard"
        size="sm"
        footer={
          <>
            <Button
              variant="secondary"
              onClick={() => setDeleteShardId(null)}
            >
              Cancel
            </Button>
            <Button
              variant="danger"
              onClick={() => {
                if (deleteShardId) {
                  deleteMutation.mutate(deleteShardId);
                }
              }}
              isLoading={deleteMutation.isPending}
            >
              Delete
            </Button>
          </>
        }
      >
        <p className="text-sm text-gray-600 dark:text-gray-300">
          Are you sure you want to delete this shard? This action cannot be undone.
        </p>
      </Modal>
    </div>
  );
}

interface CreateShardModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: CreateShardRequest) => void;
  isLoading: boolean;
}

function CreateShardModal({ isOpen, onClose, onSubmit, isLoading }: CreateShardModalProps) {
  const { data: clientApps } = useClientApps();
  const [formData, setFormData] = useState<CreateShardRequest>({
    name: '',
    client_app_id: '', // Required field
    primary_endpoint: '',
    replicas: [],
    vnode_count: 256,
  });
  const [replicaInput, setReplicaInput] = useState('');

  // Reset form when modal closes
  useEffect(() => {
    if (!isOpen) {
      setFormData({
        name: '',
        client_app_id: '',
        primary_endpoint: '',
        replicas: [],
        vnode_count: 256,
      });
      setReplicaInput('');
    }
  }, [isOpen]);

  // Auto-fill endpoint when clientApps loads and a client_app_id is already selected
  useEffect(() => {
    if (isOpen && formData.client_app_id && clientApps && clientApps.length > 0) {
      const selectedApp = clientApps.find(app => app.id === formData.client_app_id);
      if (selectedApp && !formData.primary_endpoint) {
        // Only auto-fill if endpoint is empty and we have the required fields
        const host = selectedApp.database_host?.trim();
        const port = selectedApp.database_port?.trim();
        const user = selectedApp.database_user?.trim();
        const dbName = selectedApp.database_name?.trim();
        
        if (host && port && user && dbName) {
          const password = selectedApp.database_password?.trim() || '';
          const endpoint = `postgres://${user}:${password}@${host}:${port}/${dbName}`;
          setFormData(prev => ({ ...prev, primary_endpoint: endpoint }));
        }
      }
    }
  }, [isOpen, formData.client_app_id, clientApps]);

  // Auto-fill primary_endpoint when client app is selected
  const handleClientAppChange = (clientAppId: string) => {
    if (!clientAppId) {
      setFormData(prev => ({ ...prev, client_app_id: '', primary_endpoint: '' }));
      return;
    }

    if (!clientApps || clientApps.length === 0) {
      // Client apps not loaded yet, just update the ID
      setFormData(prev => ({ ...prev, client_app_id: clientAppId }));
      return;
    }

    const selectedApp = clientApps.find(app => app.id === clientAppId);
    if (!selectedApp) {
      setFormData(prev => ({ ...prev, client_app_id: clientAppId }));
      return;
    }

    // Try to construct connection string from individual fields
    // Check if we have all required fields (host, port, user, database name)
    const host = selectedApp.database_host?.trim();
    const port = selectedApp.database_port?.trim();
    const user = selectedApp.database_user?.trim();
    const dbName = selectedApp.database_name?.trim();
    
    if (host && port && user && dbName) {
      const password = selectedApp.database_password?.trim() || '';
      const endpoint = `postgres://${user}:${password}@${host}:${port}/${dbName}`;
      setFormData(prev => ({ ...prev, client_app_id: clientAppId, primary_endpoint: endpoint }));
      return;
    }

    // If individual fields not available, just update the client_app_id
    // The endpoint will remain empty and user can fill it manually
    setFormData(prev => ({ ...prev, client_app_id: clientAppId }));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.client_app_id) {
      toast.error('Please select a client application');
      return;
    }
    onSubmit(formData);
    setFormData({
      name: '',
      client_app_id: '',
      primary_endpoint: '',
      replicas: [],
      vnode_count: 256,
    });
    setReplicaInput('');
  };

  const addReplica = () => {
    if (replicaInput.trim()) {
      setFormData({
        ...formData,
        replicas: [...formData.replicas, replicaInput.trim()],
      });
      setReplicaInput('');
    }
  };

  const removeReplica = (index: number) => {
    setFormData({
      ...formData,
      replicas: formData.replicas.filter((_, i) => i !== index),
    });
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Create Shard"
      size="lg"
      footer={
        <>
          <Button variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} isLoading={isLoading}>
            Create Shard
          </Button>
        </>
      }
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            Client Application <span className="text-red-500">*</span>
          </label>
          <select
            value={formData.client_app_id}
            onChange={(e) => handleClientAppChange(e.target.value)}
            required
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary-500 focus:border-transparent"
          >
            <option value="">Select Client Application</option>
            {clientApps?.map((app) => (
              <option key={app.id} value={app.id}>
                {app.name}
              </option>
            ))}
          </select>
          <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
            Shards must belong to a client application for proper isolation
          </p>
        </div>
        <Input
          label="Shard Name"
          value={formData.name}
          onChange={(e) => setFormData({ ...formData, name: e.target.value })}
          required
          placeholder="shard-01"
        />
        <Input
          label="Primary Database Endpoint"
          value={formData.primary_endpoint}
          onChange={(e) => setFormData({ ...formData, primary_endpoint: e.target.value })}
          required
          placeholder="postgres://user:pass@host:5432/db"
        />
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            Virtual Node Count
          </label>
          <Input
            type="number"
            value={formData.vnode_count}
            onChange={(e) => setFormData({ ...formData, vnode_count: parseInt(e.target.value) || 256 })}
            min={1}
            max={1024}
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            Replicas
          </label>
          <div className="flex space-x-2">
            <Input
              value={replicaInput}
              onChange={(e) => setReplicaInput(e.target.value)}
              placeholder="postgres://rep1:5432/db"
              onKeyPress={(e) => {
                if (e.key === 'Enter') {
                  e.preventDefault();
                  addReplica();
                }
              }}
            />
            <Button type="button" onClick={addReplica}>
              Add
            </Button>
          </div>
          {formData.replicas.length > 0 && (
            <div className="mt-2 space-y-1">
              {formData.replicas.map((replica, index) => (
                <div
                  key={index}
                  className="flex items-center justify-between bg-gray-50 dark:bg-gray-800 px-3 py-2 rounded"
                >
                  <span className="text-sm text-gray-700 dark:text-gray-300">{replica}</span>
                  <button
                    type="button"
                    onClick={() => removeReplica(index)}
                    className="text-red-600 hover:text-red-700"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      </form>
    </Modal>
  );
}

