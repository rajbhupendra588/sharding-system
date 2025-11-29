import { useState } from 'react';
import { useParams } from 'react-router-dom';
import { Plus, GitBranch, Trash2, GitMerge, Copy, CheckCircle2 } from 'lucide-react';
import { useBranches, useCreateBranch, useDeleteBranch, useMergeBranch } from '@/features/branch';
import { useDatabase } from '@/features/database';
import LoadingSpinner from '@/components/ui/LoadingSpinner';
import Button from '@/components/ui/Button';
import Input from '@/components/ui/Input';
import Modal from '@/components/ui/Modal';
import StatusBadge from '@/components/ui/StatusBadge';
import { formatRelativeTime } from '@/lib/utils';
import { toast } from 'react-hot-toast';

// Card component
const Card = ({ children, className = '' }: { children: React.ReactNode; className?: string }) => (
  <div className={`card ${className}`}>{children}</div>
);

export default function Branches() {
  const { dbName } = useParams<{ dbName: string }>();
  
  // Fallback: if dbName is not in params, try to get from database context
  const actualDbName = dbName || '';
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [isMergeModalOpen, setIsMergeModalOpen] = useState(false);
  const [selectedBranchID, setSelectedBranchID] = useState<string | null>(null);
  const [branchName, setBranchName] = useState('');
  const [copiedConnection, setCopiedConnection] = useState<string | null>(null);

  const { data: database } = useDatabase(actualDbName);
  const { data: branches, isLoading } = useBranches(actualDbName);
  const createMutation = useCreateBranch();
  const deleteMutation = useDeleteBranch();
  const mergeMutation = useMergeBranch();

  const handleCreateBranch = async () => {
    if (!actualDbName || !branchName.trim()) {
      toast.error('Branch name is required');
      return;
    }

    await createMutation.mutateAsync({
      dbName: actualDbName,
      request: { name: branchName.trim() },
    });
    setIsCreateModalOpen(false);
    setBranchName('');
  };

  const handleDeleteBranch = async () => {
    if (selectedBranchID) {
      await deleteMutation.mutateAsync(selectedBranchID);
      setIsDeleteModalOpen(false);
      setSelectedBranchID(null);
    }
  };

  const handleMergeBranch = async () => {
    if (selectedBranchID) {
      await mergeMutation.mutateAsync(selectedBranchID);
      setIsMergeModalOpen(false);
      setSelectedBranchID(null);
    }
  };

  const copyToClipboard = async (text: string, branchID: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedConnection(branchID);
      setTimeout(() => setCopiedConnection(null), 2000);
      toast.success('Copied to clipboard');
    } catch (err) {
      toast.error('Failed to copy');
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
      case 'deleting':
        return 'warning';
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

  return (
    <div className="space-y-6 p-4 sm:p-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-white">
            Database Branches
          </h1>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
            {database?.display_name || database?.name || actualDbName}
          </p>
        </div>
        <Button onClick={() => setIsCreateModalOpen(true)} className="w-full sm:w-auto">
          <Plus className="h-4 w-4 mr-2" />
          Create Branch
        </Button>
      </div>

      {/* Branches List */}
      {branches && branches.length > 0 ? (
        <Card>
          <div className="space-y-4">
            {branches.map((branch) => (
              <div
                key={branch.id}
                className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg"
              >
                <div className="flex-1">
                  <div className="flex items-center gap-2 mb-2">
                    <GitBranch className="h-5 w-5 text-blue-500" />
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                      {branch.name}
                    </h3>
                    <StatusBadge status={getStatusColor(branch.status)} />
                  </div>
                  <div className="space-y-1 text-sm text-gray-500 dark:text-gray-400">
                    <p>Branch ID: {branch.id}</p>
                    <p>Parent: {branch.parent_db_name}</p>
                    <p>Created: {formatRelativeTime(new Date(branch.created_at))}</p>
                    {branch.connection_string && (
                      <div className="flex items-center gap-2 mt-2">
                        <code className="text-xs bg-gray-100 dark:bg-gray-700 px-2 py-1 rounded flex-1 truncate">
                          {branch.connection_string}
                        </code>
                        <button
                          onClick={() => copyToClipboard(branch.connection_string, branch.id)}
                          className="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
                        >
                          {copiedConnection === branch.id ? (
                            <CheckCircle2 className="h-4 w-4 text-green-500" />
                          ) : (
                            <Copy className="h-4 w-4" />
                          )}
                        </button>
                      </div>
                    )}
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  {branch.status === 'ready' && (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => {
                        setSelectedBranchID(branch.id);
                        setIsMergeModalOpen(true);
                      }}
                    >
                      <GitMerge className="h-4 w-4 mr-2" />
                      Merge
                    </Button>
                  )}
                  <Button
                    variant="danger"
                    size="sm"
                    onClick={() => {
                      setSelectedBranchID(branch.id);
                      setIsDeleteModalOpen(true);
                    }}
                    disabled={branch.status === 'deleting'}
                  >
                    <Trash2 className="h-4 w-4 mr-2" />
                    Delete
                  </Button>
                </div>
              </div>
            ))}
          </div>
        </Card>
      ) : (
        <Card>
          <div className="text-center py-12">
            <GitBranch className="h-12 w-12 mx-auto text-gray-400 mb-4" />
            <p className="text-gray-600 dark:text-gray-400 mb-4">No branches yet</p>
            <Button onClick={() => setIsCreateModalOpen(true)}>
              <Plus className="h-4 w-4 mr-2" />
              Create Your First Branch
            </Button>
          </div>
        </Card>
      )}

      {/* Create Branch Modal */}
      <Modal
        isOpen={isCreateModalOpen}
        onClose={() => {
          setIsCreateModalOpen(false);
          setBranchName('');
        }}
        title="Create Branch"
        footer={
          <>
            <Button variant="outline" onClick={() => setIsCreateModalOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleCreateBranch}
              disabled={!branchName.trim() || createMutation.isPending}
            >
              {createMutation.isPending ? <LoadingSpinner /> : 'Create Branch'}
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <p className="text-sm text-gray-600 dark:text-gray-400">
            Create an isolated development branch from {database?.name || actualDbName}
          </p>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Branch Name
            </label>
            <Input
              value={branchName}
              onChange={(e) => setBranchName(e.target.value)}
              placeholder="e.g., dev-branch, feature-auth"
              onKeyPress={(e) => {
                if (e.key === 'Enter' && branchName.trim()) {
                  handleCreateBranch();
                }
              }}
            />
          </div>
        </div>
      </Modal>

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={isDeleteModalOpen}
        onClose={() => {
          setIsDeleteModalOpen(false);
          setSelectedBranchID(null);
        }}
        title="Delete Branch"
        footer={
          <>
            <Button variant="outline" onClick={() => setIsDeleteModalOpen(false)}>
              Cancel
            </Button>
            <Button
              variant="danger"
              onClick={handleDeleteBranch}
              disabled={deleteMutation.isPending}
            >
              {deleteMutation.isPending ? <LoadingSpinner /> : 'Delete'}
            </Button>
          </>
        }
      >
        <p className="text-gray-600 dark:text-gray-400">
          Are you sure you want to delete this branch? This action cannot be undone.
        </p>
      </Modal>

      {/* Merge Confirmation Modal */}
      <Modal
        isOpen={isMergeModalOpen}
        onClose={() => {
          setIsMergeModalOpen(false);
          setSelectedBranchID(null);
        }}
        title="Merge Branch"
        footer={
          <>
            <Button variant="outline" onClick={() => setIsMergeModalOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleMergeBranch}
              disabled={mergeMutation.isPending}
            >
              {mergeMutation.isPending ? <LoadingSpinner /> : 'Merge'}
            </Button>
          </>
        }
      >
        <p className="text-gray-600 dark:text-gray-400">
          Merge schema changes from this branch into the parent database? This will apply all schema changes made in the branch.
        </p>
      </Modal>
    </div>
  );
}

