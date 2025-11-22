import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { Plus, Search, Trash2, ExternalLink } from 'lucide-react';
import { apiClient } from '@/lib/api-client';
import { toast } from 'react-hot-toast';
import LoadingSpinner from '@/components/ui/LoadingSpinner';
import Button from '@/components/ui/Button';
import Input from '@/components/ui/Input';
import Modal from '@/components/ui/Modal';
import StatusBadge from '@/components/ui/StatusBadge';
import { Table, TableHead, TableHeader, TableBody, TableRow, TableCell } from '@/components/ui/Table';
import { formatRelativeTime } from '@/lib/utils';
import type { CreateShardRequest } from '@/types';

export default function Shards() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [searchTerm, setSearchTerm] = useState('');
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [deleteShardId, setDeleteShardId] = useState<string | null>(null);

  const { data: shards, isLoading } = useQuery({
    queryKey: ['shards'],
    queryFn: () => apiClient.listShards(),
    refetchInterval: 10000,
  });

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

  const filteredShards = shards?.filter((shard) =>
    shard.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    shard.id.toLowerCase().includes(searchTerm.toLowerCase())
  ) || [];

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
              <TableHeader>Status</TableHeader>
              <TableHeader>Primary Endpoint</TableHeader>
              <TableHeader>Replicas</TableHeader>
              <TableHeader>Created</TableHeader>
              <TableHeader>Actions</TableHeader>
            </TableHead>
            <TableBody>
              {filteredShards.map((shard) => (
                <TableRow
                  key={shard.id}
                  onClick={() => navigate(`/shards/${shard.id}`)}
                >
                  <TableCell>
                    <div>
                      <div className="font-medium text-gray-900 dark:text-white">{shard.name}</div>
                      <div className="text-sm text-gray-500 dark:text-gray-400">{shard.id}</div>
                    </div>
                  </TableCell>
                  <TableCell>
                    <StatusBadge status={shard.status} />
                  </TableCell>
                  <TableCell>
                    <div className="text-sm text-gray-900 dark:text-gray-300 max-w-xs truncate">
                      {shard.primary_endpoint}
                    </div>
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
              ))}
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
  const [formData, setFormData] = useState<CreateShardRequest>({
    name: '',
    primary_endpoint: '',
    replicas: [],
    vnode_count: 256,
  });
  const [replicaInput, setReplicaInput] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(formData);
    setFormData({
      name: '',
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
        <Input
          label="Shard Name"
          value={formData.name}
          onChange={(e) => setFormData({ ...formData, name: e.target.value })}
          required
          placeholder="shard-01"
        />
        <Input
          label="Primary Endpoint"
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

