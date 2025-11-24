import { useState } from 'react';
import * as React from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { Plus, GitBranch, ArrowRight, Users } from 'lucide-react';
import { apiClient } from '@/lib/api-client';
import { toast } from 'react-hot-toast';
import Button from '@/components/ui/Button';
import Modal from '@/components/ui/Modal';
import { useClientApps } from '@/features/clientApp';
import type { SplitRequest, MergeRequest, CreateShardRequest } from '@/types';

export default function Resharding() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [isSplitModalOpen, setIsSplitModalOpen] = useState(false);
  const [isMergeModalOpen, setIsMergeModalOpen] = useState(false);

  // Note: In a real implementation, you'd have an endpoint to list all reshard jobs
  // For now, we'll show a placeholder
  const { data: shards } = useQuery({
    queryKey: ['shards'],
    queryFn: () => apiClient.listShards(),
  });

  const splitMutation = useMutation({
    mutationFn: (request: SplitRequest) => apiClient.splitShard(request),
    onSuccess: (job) => {
      queryClient.invalidateQueries({ queryKey: ['shards'] });
      setIsSplitModalOpen(false);
      toast.success('Split operation started');
      navigate(`/resharding/jobs/${job.id}`);
    },
    onError: (error: { message: string }) => {
      toast.error(`Failed to start split: ${error.message}`);
    },
  });

  const mergeMutation = useMutation({
    mutationFn: (request: MergeRequest) => apiClient.mergeShards(request),
    onSuccess: (job) => {
      queryClient.invalidateQueries({ queryKey: ['shards'] });
      setIsMergeModalOpen(false);
      toast.success('Merge operation started');
      navigate(`/resharding/jobs/${job.id}`);
    },
    onError: (error: { message: string }) => {
      toast.error(`Failed to start merge: ${error.message}`);
    },
  });

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white">Resharding</h1>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
            Split or merge shards to scale your database
          </p>
        </div>
        <div className="flex items-center space-x-2">
          <Button variant="secondary" onClick={() => setIsMergeModalOpen(true)}>
            <GitBranch className="h-4 w-4 mr-2" />
            Merge Shards
          </Button>
          <Button onClick={() => setIsSplitModalOpen(true)}>
            <Plus className="h-4 w-4 mr-2" />
            Split Shard
          </Button>
        </div>
      </div>

      {/* Info Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="card">
          <div className="flex items-center mb-4">
            <div className="flex-shrink-0 bg-blue-50 rounded-lg p-3">
              <GitBranch className="h-6 w-6 text-blue-600" />
            </div>
            <div className="ml-4">
              <h3 className="text-lg font-medium text-gray-900 dark:text-white">Split Shard</h3>
              <p className="text-sm text-gray-500 dark:text-gray-400">Divide a large shard into multiple smaller shards</p>
            </div>
          </div>
          <ul className="space-y-2 text-sm text-gray-600 dark:text-gray-300">
            <li>• Reduces load on individual shards</li>
            <li>• Enables horizontal scaling</li>
            <li>• Minimal downtime during cutover</li>
          </ul>
        </div>

        <div className="card">
          <div className="flex items-center mb-4">
            <div className="flex-shrink-0 bg-purple-50 rounded-lg p-3">
              <ArrowRight className="h-6 w-6 text-purple-600" />
            </div>
            <div className="ml-4">
              <h3 className="text-lg font-medium text-gray-900 dark:text-white">Merge Shards</h3>
              <p className="text-sm text-gray-500 dark:text-gray-400">Combine multiple small shards into one</p>
            </div>
          </div>
          <ul className="space-y-2 text-sm text-gray-600 dark:text-gray-300">
            <li>• Consolidates resources</li>
            <li>• Reduces operational overhead</li>
            <li>• Optimizes for smaller workloads</li>
          </ul>
        </div>
      </div>

      {/* Active Jobs */}
      <div className="card">
        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">Active Resharding Jobs</h2>
        <div className="text-center py-12 text-gray-500 dark:text-gray-400">
          <p>Job tracking will be available once jobs are created</p>
          <p className="text-sm mt-1">Start a split or merge operation to see job progress</p>
        </div>
      </div>

      {/* Split Modal */}
      <SplitShardModal
        isOpen={isSplitModalOpen}
        onClose={() => setIsSplitModalOpen(false)}
        onSubmit={(data) => splitMutation.mutate(data)}
        isLoading={splitMutation.isPending}
        shards={shards || []}
      />

      {/* Merge Modal */}
      <MergeShardsModal
        isOpen={isMergeModalOpen}
        onClose={() => setIsMergeModalOpen(false)}
        onSubmit={(data) => mergeMutation.mutate(data)}
        isLoading={mergeMutation.isPending}
        shards={shards || []}
      />
    </div>
  );
}

interface SplitShardModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: SplitRequest) => void;
  isLoading: boolean;
  shards: Array<{ id: string; name: string; client_app_id?: string }>;
}

function SplitShardModal({ isOpen, onClose, onSubmit, isLoading, shards }: SplitShardModalProps) {
  const { data: clientApps } = useClientApps();
  const clientAppMap = new Map(clientApps?.map(app => [app.id, app.name]) || []);
  const [sourceShardId, setSourceShardId] = useState('');
  const [targetShards, setTargetShards] = useState<CreateShardRequest[]>([
    { name: '', client_app_id: '', primary_endpoint: '', replicas: [], vnode_count: 256 },
  ]);

  // Get selected source shard to determine client app
  const sourceShard = shards.find(s => s.id === sourceShardId);
  const sourceClientAppId = sourceShard?.client_app_id || '';

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!sourceShardId || targetShards.length === 0) {
      toast.error('Please fill in all required fields');
      return;
    }
    if (!sourceClientAppId) {
      toast.error('Source shard must belong to a client application');
      return;
    }
    // Ensure all target shards have the same client_app_id as source
    const validatedTargets = targetShards.map(t => ({
      ...t,
      client_app_id: sourceClientAppId, // Inherit from source shard
    }));
    onSubmit({
      source_shard_id: sourceShardId,
      target_shards: validatedTargets,
    });
  };

  const addTargetShard = () => {
    setTargetShards([...targetShards, {
      name: '',
      client_app_id: sourceClientAppId, // Inherit client app ID
      primary_endpoint: '',
      replicas: [],
      vnode_count: 256
    }]);
  };

  // Update target shards when source changes
  React.useEffect(() => {
    if (sourceClientAppId) {
      setTargetShards(prev => prev.map(t => ({ ...t, client_app_id: sourceClientAppId })));
    }
  }, [sourceClientAppId]);

  const updateTargetShard = (index: number, field: keyof CreateShardRequest, value: unknown) => {
    const updated = [...targetShards];
    updated[index] = { ...updated[index], [field]: value };
    setTargetShards(updated);
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Split Shard"
      size="xl"
      footer={
        <>
          <Button variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} isLoading={isLoading}>
            Start Split
          </Button>
        </>
      }
    >
      <form onSubmit={handleSubmit} className="space-y-6">
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            Source Shard
          </label>
          <select
            value={sourceShardId}
            onChange={(e) => setSourceShardId(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 bg-white dark:bg-gray-800 text-gray-900 dark:text-white"
            required
          >
            <option value="">Select a shard</option>
            {shards.map((shard) => {
              const clientAppName = shard.client_app_id ? clientAppMap.get(shard.client_app_id) : 'Unknown';
              return (
                <option key={shard.id} value={shard.id}>
                  {shard.name} - {clientAppName} ({shard.id.substring(0, 8)}...)
                </option>
              );
            })}
          </select>
          {sourceShard && sourceShard.client_app_id && (
            <div className="mt-2 p-2 bg-blue-50 dark:bg-blue-900/20 rounded text-sm">
              <div className="flex items-center gap-2 text-blue-700 dark:text-blue-300">
                <Users className="h-4 w-4" />
                <span>Client Application: <strong>{clientAppMap.get(sourceShard.client_app_id) || sourceShard.client_app_id}</strong></span>
              </div>
              <p className="text-xs text-blue-600 dark:text-blue-400 mt-1">
                Target shards will belong to the same client application
              </p>
            </div>
          )}
        </div>

        <div>
          <div className="flex items-center justify-between mb-2">
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
              Target Shards
            </label>
            <Button type="button" size="sm" onClick={addTargetShard}>
              Add Target Shard
            </Button>
          </div>
          <div className="space-y-4">
            {targetShards.map((target, index) => (
              <div key={index} className="border border-gray-200 dark:border-gray-700 rounded-lg p-4 space-y-3">
                <div className="flex items-center justify-between">
                  <h4 className="font-medium text-gray-900 dark:text-white">Target Shard {index + 1}</h4>
                  {targetShards.length > 1 && (
                    <button
                      type="button"
                      onClick={() => setTargetShards(targetShards.filter((_, i) => i !== index))}
                      className="text-red-600 hover:text-red-700 text-sm"
                    >
                      Remove
                    </button>
                  )}
                </div>
                <input
                  type="text"
                  placeholder="Shard name"
                  value={target.name}
                  onChange={(e) => updateTargetShard(index, 'name', e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white"
                  required
                />
                <input
                  type="text"
                  placeholder="Primary endpoint (postgres://...)"
                  value={target.primary_endpoint}
                  onChange={(e) => updateTargetShard(index, 'primary_endpoint', e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg font-mono text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-white"
                  required
                />
                <input
                  type="number"
                  placeholder="Virtual node count"
                  value={target.vnode_count}
                  onChange={(e) => updateTargetShard(index, 'vnode_count', parseInt(e.target.value) || 256)}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white"
                  min={1}
                  max={1024}
                />
              </div>
            ))}
          </div>
        </div>
      </form>
    </Modal>
  );
}

interface MergeShardsModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: MergeRequest) => void;
  isLoading: boolean;
  shards: Array<{ id: string; name: string; client_app_id?: string }>;
}

function MergeShardsModal({ isOpen, onClose, onSubmit, isLoading, shards }: MergeShardsModalProps) {
  const { data: clientApps } = useClientApps();
  const clientAppMap = new Map(clientApps?.map(app => [app.id, app.name]) || []);
  const [sourceShardIds, setSourceShardIds] = useState<string[]>([]);
  const [targetShard, setTargetShard] = useState<CreateShardRequest>({
    name: '',
    client_app_id: '',
    primary_endpoint: '',
    replicas: [],
    vnode_count: 256,
  });

  // Get client app ID from selected source shards (must all be same)
  const selectedShards = shards.filter(s => sourceShardIds.includes(s.id));
  const sourceClientAppId = selectedShards.length > 0 && selectedShards.every(s => s.client_app_id === selectedShards[0].client_app_id)
    ? selectedShards[0].client_app_id
    : '';

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (sourceShardIds.length < 2) {
      toast.error('Please select at least 2 source shards');
      return;
    }
    if (!sourceClientAppId) {
      toast.error('All source shards must belong to the same client application');
      return;
    }
    onSubmit({
      source_shard_ids: sourceShardIds,
      target_shard: {
        ...targetShard,
        client_app_id: sourceClientAppId, // Inherit from source shards
      },
    });
  };

  const toggleSourceShard = (shardId: string) => {
    if (sourceShardIds.includes(shardId)) {
      setSourceShardIds(sourceShardIds.filter((id) => id !== shardId));
    } else {
      setSourceShardIds([...sourceShardIds, shardId]);
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Merge Shards"
      size="xl"
      footer={
        <>
          <Button variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} isLoading={isLoading}>
            Start Merge
          </Button>
        </>
      }
    >
      <form onSubmit={handleSubmit} className="space-y-6">
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Source Shards (select at least 2)
          </label>
          <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-4 max-h-48 overflow-y-auto space-y-2">
            {shards.map((shard) => {
              const clientAppName = shard.client_app_id ? clientAppMap.get(shard.client_app_id) : 'Unknown';
              const isSelected = sourceShardIds.includes(shard.id);
              const isCompatible = !sourceShardIds.length ||
                (shard.client_app_id && sourceShardIds.every(id => {
                  const otherShard = shards.find(s => s.id === id);
                  return otherShard?.client_app_id === shard.client_app_id;
                }));

              return (
                <label
                  key={shard.id}
                  className={`flex items-start space-x-2 cursor-pointer p-2 rounded ${isSelected ? 'bg-primary-50 dark:bg-primary-900/20' :
                    !isCompatible ? 'opacity-50' : 'hover:bg-gray-50 dark:hover:bg-gray-800'
                    }`}
                >
                  <input
                    type="checkbox"
                    checked={isSelected}
                    onChange={() => toggleSourceShard(shard.id)}
                    disabled={!isCompatible && !isSelected}
                    className="rounded border-gray-300 text-primary-600 focus:ring-primary-500 bg-white dark:bg-gray-800 mt-0.5"
                  />
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium text-gray-900 dark:text-white">{shard.name}</span>
                      <span className="text-xs text-gray-500 dark:text-gray-400 font-mono">({shard.id.substring(0, 8)}...)</span>
                    </div>
                    <div className="flex items-center gap-1 mt-1">
                      <Users className="h-3 w-3 text-gray-400" />
                      <span className="text-xs text-gray-600 dark:text-gray-400">{clientAppName || 'No client app'}</span>
                    </div>
                    {!isCompatible && !isSelected && (
                      <p className="text-xs text-red-600 dark:text-red-400 mt-1">
                        Must belong to same client app as selected shards
                      </p>
                    )}
                  </div>
                </label>
              );
            })}
          </div>
          {sourceClientAppId && (
            <div className="mt-2 p-2 bg-blue-50 dark:bg-blue-900/20 rounded text-sm">
              <div className="flex items-center gap-2 text-blue-700 dark:text-blue-300">
                <Users className="h-4 w-4" />
                <span>Target shard will belong to: <strong>{clientAppMap.get(sourceClientAppId) || sourceClientAppId}</strong></span>
              </div>
            </div>
          )}
        </div>

        <div className="border-t border-gray-200 dark:border-gray-700 pt-4">
          <h4 className="font-medium text-gray-900 dark:text-white mb-3">Target Shard Configuration</h4>
          {sourceClientAppId && (
            <div className="mb-3 p-2 bg-gray-50 dark:bg-gray-800 rounded text-xs text-gray-600 dark:text-gray-400">
              Client Application: <strong className="text-gray-900 dark:text-white">{clientAppMap.get(sourceClientAppId) || sourceClientAppId}</strong> (inherited from source shards)
            </div>
          )}
          <div className="space-y-3">
            <input
              type="text"
              placeholder="Shard name"
              value={targetShard.name}
              onChange={(e) => setTargetShard({ ...targetShard, name: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white"
              required
            />
            <input
              type="text"
              placeholder="Primary endpoint (postgres://...)"
              value={targetShard.primary_endpoint}
              onChange={(e) => setTargetShard({ ...targetShard, primary_endpoint: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg font-mono text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-white"
              required
            />
            <input
              type="number"
              placeholder="Virtual node count"
              value={targetShard.vnode_count}
              onChange={(e) => setTargetShard({ ...targetShard, vnode_count: parseInt(e.target.value) || 256 })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white"
              min={1}
              max={1024}
            />
          </div>
        </div>
      </form>
    </Modal>
  );
}

