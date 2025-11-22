import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { ArrowLeft, Trash2 } from 'lucide-react';
import { apiClient } from '@/lib/api-client';
import { toast } from 'react-hot-toast';
import LoadingSpinner from '@/components/ui/LoadingSpinner';
import Button from '@/components/ui/Button';
import StatusBadge from '@/components/ui/StatusBadge';
import Modal from '@/components/ui/Modal';
import Input from '@/components/ui/Input';
import { formatDate, formatRelativeTime } from '@/lib/utils';
import { useState } from 'react';
import type { PromoteReplicaRequest } from '@/types';

export default function ShardDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [isPromoteModalOpen, setIsPromoteModalOpen] = useState(false);
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [replicaEndpoint, setReplicaEndpoint] = useState('');

  const { data: shard, isLoading } = useQuery({
    queryKey: ['shard', id],
    queryFn: () => apiClient.getShard(id!),
    enabled: !!id,
    refetchInterval: 10000,
  });

  const deleteMutation = useMutation({
    mutationFn: () => apiClient.deleteShard(id!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['shards'] });
      navigate('/shards');
      toast.success('Shard deleted successfully');
    },
    onError: (error: { message: string }) => {
      toast.error(`Failed to delete shard: ${error.message}`);
    },
  });

  const promoteMutation = useMutation({
    mutationFn: (data: PromoteReplicaRequest) => apiClient.promoteReplica(id!, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['shard', id] });
      setIsPromoteModalOpen(false);
      setReplicaEndpoint('');
      toast.success('Replica promoted successfully');
    },
    onError: (error: { message: string }) => {
      toast.error(`Failed to promote replica: ${error.message}`);
    },
  });

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <LoadingSpinner size="lg" />
      </div>
    );
  }

  if (!shard) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500">Shard not found</p>
        <Button onClick={() => navigate('/shards')} className="mt-4">
          Back to Shards
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <Button variant="secondary" onClick={() => navigate('/shards')}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back
          </Button>
          <div>
            <h1 className="text-3xl font-bold text-gray-900 dark:text-white">{shard.name}</h1>
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">{shard.id}</p>
          </div>
        </div>
        <div className="flex items-center space-x-2">
          <StatusBadge status={shard.status} />
          <Button
            variant="danger"
            onClick={() => setIsDeleteModalOpen(true)}
          >
            <Trash2 className="h-4 w-4 mr-2" />
            Delete
          </Button>
        </div>
      </div>

      {/* Details Grid */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        {/* Basic Information */}
        <div className="card">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Basic Information</h2>
          <dl className="space-y-3">
            <div>
              <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Status</dt>
              <dd className="mt-1">
                <StatusBadge status={shard.status} />
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Created</dt>
              <dd className="mt-1 text-sm text-gray-900 dark:text-gray-300">{formatDate(shard.created_at)}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Last Updated</dt>
              <dd className="mt-1 text-sm text-gray-900 dark:text-gray-300">{formatRelativeTime(shard.updated_at)}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Version</dt>
              <dd className="mt-1 text-sm text-gray-900 dark:text-gray-300">{shard.version}</dd>
            </div>
            {shard.hash_range_start !== undefined && shard.hash_range_end !== undefined && (
              <>
                <div>
                  <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Hash Range Start</dt>
                  <dd className="mt-1 text-sm text-gray-900 dark:text-gray-300 font-mono">{shard.hash_range_start}</dd>
                </div>
                <div>
                  <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Hash Range End</dt>
                  <dd className="mt-1 text-sm text-gray-900 dark:text-gray-300 font-mono">{shard.hash_range_end}</dd>
                </div>
              </>
            )}
          </dl>
        </div>

        {/* Endpoints */}
        <div className="card">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Endpoints</h2>
          <div className="space-y-4">
            <div>
              <dt className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">Primary Endpoint</dt>
              <dd className="text-sm text-gray-900 dark:text-gray-300 font-mono bg-gray-50 dark:bg-gray-800 p-2 rounded break-all">
                {shard.primary_endpoint}
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">
                Replicas ({shard.replicas?.length || 0})
              </dt>
              {shard.replicas?.length > 0 ? (
                <div className="space-y-2">
                  {shard.replicas.map((replica, index) => (
                    <div
                      key={index}
                      className="flex items-center justify-between bg-gray-50 dark:bg-gray-800 p-2 rounded"
                    >
                      <span className="text-sm text-gray-900 dark:text-gray-300 font-mono break-all flex-1">
                        {replica}
                      </span>
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => {
                          setReplicaEndpoint(replica);
                          setIsPromoteModalOpen(true);
                        }}
                      >
                        Promote
                      </Button>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-sm text-gray-500 dark:text-gray-400">No replicas configured</p>
              )}
            </div>
          </div>
        </div>

        {/* Virtual Nodes */}
        {shard.vnodes && shard.vnodes.length > 0 && (
          <div className="card lg:col-span-2">
            <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              Virtual Nodes ({shard.vnodes.length})
            </h2>
            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-2">
              {shard.vnodes.slice(0, 20).map((vnode) => (
                <div
                  key={vnode.id}
                  className="bg-gray-50 dark:bg-gray-800 p-2 rounded text-sm font-mono"
                >
                  <div className="text-gray-600 dark:text-gray-400">ID: {vnode.id}</div>
                  <div className="text-gray-900 dark:text-gray-300">Hash: {vnode.hash}</div>
                </div>
              ))}
              {shard.vnodes.length > 20 && (
                <div className="bg-gray-50 dark:bg-gray-800 p-2 rounded text-sm text-gray-600 dark:text-gray-400 flex items-center justify-center">
                  +{shard.vnodes.length - 20} more
                </div>
              )}
            </div>
          </div>
        )}
      </div>

      {/* Promote Replica Modal */}
      <Modal
        isOpen={isPromoteModalOpen}
        onClose={() => {
          setIsPromoteModalOpen(false);
          setReplicaEndpoint('');
        }}
        title="Promote Replica"
        size="md"
        footer={
          <>
            <Button
              variant="secondary"
              onClick={() => {
                setIsPromoteModalOpen(false);
                setReplicaEndpoint('');
              }}
            >
              Cancel
            </Button>
            <Button
              onClick={() => {
                promoteMutation.mutate({ replica_endpoint: replicaEndpoint });
              }}
              isLoading={promoteMutation.isPending}
            >
              Promote
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <p className="text-sm text-gray-600 dark:text-gray-300">
            Are you sure you want to promote this replica to primary? This will update the shard configuration.
          </p>
          <Input
            label="Replica Endpoint"
            value={replicaEndpoint}
            onChange={(e) => setReplicaEndpoint(e.target.value)}
            readOnly
          />
        </div>
      </Modal>

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={isDeleteModalOpen}
        onClose={() => setIsDeleteModalOpen(false)}
        title="Delete Shard"
        size="sm"
        footer={
          <>
            <Button variant="secondary" onClick={() => setIsDeleteModalOpen(false)}>
              Cancel
            </Button>
            <Button
              variant="danger"
              onClick={() => deleteMutation.mutate()}
              isLoading={deleteMutation.isPending}
            >
              Delete
            </Button>
          </>
        }
      >
        <p className="text-sm text-gray-600 dark:text-gray-300">
          Are you sure you want to delete shard <strong>{shard.name}</strong>? This action cannot be undone.
        </p>
      </Modal>
    </div>
  );
}

