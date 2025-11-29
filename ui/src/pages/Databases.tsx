import { useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { Plus, Search, Database as DatabaseIcon, Copy, CheckCircle2 } from 'lucide-react';
import { Link } from 'react-router-dom';
import { toast } from 'react-hot-toast';
import LoadingSpinner from '@/components/ui/LoadingSpinner';
import Button from '@/components/ui/Button';
import Input from '@/components/ui/Input';
import Modal from '@/components/ui/Modal';
import StatusBadge from '@/components/ui/StatusBadge';
import { Table, TableHead, TableHeader, TableBody, TableRow, TableCell } from '@/components/ui/Table';
import { formatRelativeTime } from '@/lib/utils';
import { useDatabases } from '@/features/database';
import DatabaseWizard from '@/components/database/DatabaseWizard';

export default function Databases() {
  const queryClient = useQueryClient();
  const [searchTerm, setSearchTerm] = useState('');
  const [isWizardOpen, setIsWizardOpen] = useState(false);
  const [copiedId, setCopiedId] = useState<string | null>(null);

  const { data: databases, isLoading } = useDatabases();

  const filteredDatabases = databases?.filter((db) => {
    const matchesSearch =
      db.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      db.id.toLowerCase().includes(searchTerm.toLowerCase()) ||
      (db.display_name && db.display_name.toLowerCase().includes(searchTerm.toLowerCase()));
    return matchesSearch;
  }) || [];

  const copyToClipboard = async (text: string, id: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedId(id);
      setTimeout(() => setCopiedId(null), 2000);
      toast.success('Copied to clipboard');
    } catch (err) {
      toast.error('Failed to copy');
    }
  };

  const getStatusDisplay = (status: string) => {
    // Normalize status values from API to display-friendly format
    if (!status || status === '0' || status === '') {
      return 'pending';
    }
    // Map common status variations
    const statusLower = status.toLowerCase();
    if (statusLower === 'success' || statusLower === 'active') {
      return 'ready';
    }
    return statusLower;
  };

  return (
    <div className="space-y-6 p-4 sm:p-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <h1 className="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-white">Databases</h1>
        <Button onClick={() => setIsWizardOpen(true)} className="w-full sm:w-auto">
          <Plus className="h-4 w-4 mr-2" />
          Create Database
        </Button>
      </div>

      {/* Search */}
      <div className="card">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-gray-400" />
          <Input
            type="text"
            placeholder="Search databases..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-10 w-full"
          />
        </div>
      </div>

      {/* Databases Table */}
      {isLoading ? (
        <div className="flex justify-center py-12">
          <LoadingSpinner />
        </div>
      ) : filteredDatabases.length === 0 ? (
        <div className="card text-center py-12">
          <DatabaseIcon className="h-12 w-12 mx-auto text-gray-400 mb-4" />
          <p className="text-gray-600 dark:text-gray-400 mb-4">
            {searchTerm ? 'No databases found matching your search' : 'No databases yet'}
          </p>
          {!searchTerm && (
            <Button onClick={() => setIsWizardOpen(true)}>
              <Plus className="h-4 w-4 mr-2" />
              Create Your First Database
            </Button>
          )}
        </div>
      ) : (
        <div className="card overflow-x-auto">
          <Table>
            <TableHead>
              <TableHeader>Name</TableHeader>
              <TableHeader>Template</TableHeader>
              <TableHeader>Status</TableHeader>
              <TableHeader>Shards</TableHeader>
              <TableHeader>Connection</TableHeader>
              <TableHeader>Created</TableHeader>
              <TableHeader className="text-right">Actions</TableHeader>
            </TableHead>
            <TableBody>
              {filteredDatabases.map((db) => (
                <TableRow key={db.id}>
                  <TableCell>
                    <div>
                      <div className="font-medium text-gray-900 dark:text-white">
                        {db.display_name || db.name}
                      </div>
                      <div className="text-sm text-gray-500 dark:text-gray-400">{db.id}</div>
                    </div>
                  </TableCell>
                  <TableCell>
                    <span className="px-2 py-1 text-xs font-medium rounded bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200">
                      {db.template}
                    </span>
                  </TableCell>
                  <TableCell>
                    <StatusBadge status={getStatusDisplay(db.status)} />
                  </TableCell>
                  <TableCell>{db.shard_ids?.length ?? 0}</TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <code className="text-xs bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded max-w-[200px] truncate">
                        {db.connection_string}
                      </code>
                      <button
                        onClick={() => copyToClipboard(db.connection_string, `conn-${db.id}`)}
                        className="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
                      >
                        {copiedId === `conn-${db.id}` ? (
                          <CheckCircle2 className="h-4 w-4 text-green-500" />
                        ) : (
                          <Copy className="h-4 w-4" />
                        )}
                      </button>
                    </div>
                  </TableCell>
                  <TableCell className="text-sm text-gray-500 dark:text-gray-400">
                    {formatRelativeTime(new Date(db.created_at))}
                  </TableCell>
                  <TableCell className="text-right">
                    <Link to={`/databases/${db.id}`}>
                      <Button variant="outline" size="sm">
                        View
                      </Button>
                    </Link>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      {/* Database Creation Wizard */}
      <Modal
        isOpen={isWizardOpen}
        onClose={() => setIsWizardOpen(false)}
        size="xl"
        title="Create Database"
      >
        <DatabaseWizard
          onSuccess={() => {
            setIsWizardOpen(false);
            queryClient.invalidateQueries({ queryKey: ['databases'] });
          }}
          onCancel={() => setIsWizardOpen(false)}
        />
      </Modal>
    </div>
  );
}

