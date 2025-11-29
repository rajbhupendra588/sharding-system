import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Database,
  Plus,
  Search,
  Table2,
  Key,
  Trash2,
  Edit2,
  RefreshCw,
  Zap,
  Hash,
  Radio,
  TestTube2,
  CheckCircle2,
  XCircle,
  ArrowRight,
  Code2,
  Copy,
  Sparkles,
} from 'lucide-react';
import Button from '@/components/ui/Button';
import LoadingSpinner from '@/components/ui/LoadingSpinner';
import Modal from '@/components/ui/Modal';
import { toast } from 'react-hot-toast';

// Proxy Admin URL - configurable
const PROXY_ADMIN_URL = import.meta.env.VITE_PROXY_ADMIN_URL || 'http://localhost:8082';

interface ShardingRule {
  table: string;
  shard_key: string;
  strategy: 'hash' | 'range' | 'broadcast';
  description?: string;
}

interface DatabaseConfig {
  id: string;
  name: string;
  database: string;
  sharding_rules: ShardingRule[];
  default_shard?: string;
}

interface QueryTestResult {
  query: string;
  database: string;
  table: string;
  shard_key: string;
  shard_value: string;
  strategy: string;
  routing: string;
  target_shard?: string;
  target_endpoint?: string;
  parsed?: {
    Type: string;
    Table: string;
    ShardKey: string;
    ShardValue: string;
    CanRoute: boolean;
    IsMultiShard: boolean;
  };
  error?: string;
}

// Fetch all sharding rules
async function fetchRules(): Promise<Record<string, DatabaseConfig>> {
  const response = await fetch(`${PROXY_ADMIN_URL}/api/v1/rules`);
  if (!response.ok) {
    if (response.status === 404) {
      return {};
    }
    throw new Error('Failed to fetch sharding rules');
  }
  return response.json();
}

// Create/update rules for a database
async function createRules(database: string, data: { name: string; sharding_rules: ShardingRule[] }) {
  const response = await fetch(`${PROXY_ADMIN_URL}/api/v1/rules/${database}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Failed to create rules' }));
    throw new Error(error.message || 'Failed to create rules');
  }
  return response.json();
}

// Update a single rule (currently unused, but kept for future use)
// async function updateRule(database: string, table: string, rule: Partial<ShardingRule>) {
//   const response = await fetch(`${PROXY_ADMIN_URL}/api/v1/rules/${database}/${table}`, {
//     method: 'PUT',
//     headers: { 'Content-Type': 'application/json' },
//     body: JSON.stringify(rule),
//   });
//   if (!response.ok) {
//     throw new Error('Failed to update rule');
//   }
//   return response.json();
// }

// Delete a rule
async function deleteRule(database: string, table: string) {
  const response = await fetch(`${PROXY_ADMIN_URL}/api/v1/rules/${database}/${table}`, {
    method: 'DELETE',
  });
  if (!response.ok) {
    throw new Error('Failed to delete rule');
  }
}

// Test query routing
async function testQuery(database: string, query: string): Promise<QueryTestResult> {
  const response = await fetch(`${PROXY_ADMIN_URL}/api/v1/query`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ database, query }),
  });
  if (!response.ok) {
    throw new Error('Failed to test query');
  }
  return response.json();
}

export default function ShardingRules() {
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedDatabase, setSelectedDatabase] = useState<string | null>(null);
  const [showAddDatabaseModal, setShowAddDatabaseModal] = useState(false);
  const [showAddRuleModal, setShowAddRuleModal] = useState(false);
  const [showTestQueryModal, setShowTestQueryModal] = useState(false);
  const [editingRule, setEditingRule] = useState<{ database: string; rule: ShardingRule } | null>(null);

  // Form states
  const [newDatabase, setNewDatabase] = useState({ name: '', database: '' });
  const [newRule, setNewRule] = useState<ShardingRule>({ table: '', shard_key: '', strategy: 'hash', description: '' });
  const [testQueryInput, setTestQueryInput] = useState({ database: '', query: '' });
  const [testResult, setTestResult] = useState<QueryTestResult | null>(null);

  const { data: rules, isLoading, error, refetch } = useQuery({
    queryKey: ['shardingRules'],
    queryFn: fetchRules,
    retry: 1,
  });

  const createRulesMutation = useMutation({
    mutationFn: ({ database, data }: { database: string; data: { name: string; sharding_rules: ShardingRule[] } }) =>
      createRules(database, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['shardingRules'] });
      setShowAddDatabaseModal(false);
      setNewDatabase({ name: '', database: '' });
      toast.success('Database sharding configuration created');
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  const addRuleMutation = useMutation({
    mutationFn: async ({ database, rule }: { database: string; rule: ShardingRule }) => {
      const existing = rules?.[database];
      const updatedRules = existing
        ? [...existing.sharding_rules.filter(r => r.table !== rule.table), rule]
        : [rule];
      return createRules(database, { name: existing?.name || database, sharding_rules: updatedRules });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['shardingRules'] });
      setShowAddRuleModal(false);
      setEditingRule(null);
      setNewRule({ table: '', shard_key: '', strategy: 'hash', description: '' });
      toast.success('Sharding rule saved');
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  const deleteRuleMutation = useMutation({
    mutationFn: ({ database, table }: { database: string; table: string }) => deleteRule(database, table),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['shardingRules'] });
      toast.success('Sharding rule deleted');
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  const testQueryMutation = useMutation({
    mutationFn: ({ database, query }: { database: string; query: string }) => testQuery(database, query),
    onSuccess: (data) => {
      setTestResult(data);
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
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
          <XCircle className="mx-auto h-12 w-12 text-red-400 mb-4" />
          <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">
            Proxy Not Available
          </h3>
          <p className="text-gray-500 dark:text-gray-400 mb-4">
            The sharding proxy admin API is not reachable at {PROXY_ADMIN_URL}
          </p>
          <p className="text-sm text-gray-500 dark:text-gray-400 mb-6">
            Make sure the proxy server is running: <code className="px-2 py-1 bg-gray-100 dark:bg-gray-800 rounded">make run-proxy</code>
          </p>
          <Button onClick={() => refetch()}>
            <RefreshCw className="h-4 w-4 mr-2" />
            Retry
          </Button>
        </div>
      </div>
    );
  }

  const databases = Object.entries(rules || {});
  const filteredDatabases = databases.filter(([key, config]) =>
    key.toLowerCase().includes(searchQuery.toLowerCase()) ||
    config.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const copyConnectionString = () => {
    const connString = `jdbc:postgresql://localhost:5432/${selectedDatabase || 'your_database'}`;
    navigator.clipboard.writeText(connString);
    toast.success('Connection string copied!');
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white flex items-center gap-3">
            <Sparkles className="h-8 w-8 text-yellow-500" />
            Zero-Code Sharding Rules
          </h1>
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
            Configure sharding rules visually - no code changes required in your application
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="secondary" onClick={() => setShowTestQueryModal(true)}>
            <TestTube2 className="h-4 w-4 mr-2" />
            Test Query
          </Button>
          <Button onClick={() => setShowAddDatabaseModal(true)}>
            <Plus className="h-4 w-4 mr-2" />
            Add Database
          </Button>
        </div>
      </div>

      {/* Zero-Code Info Banner */}
      <div className="bg-gradient-to-r from-emerald-50 to-teal-50 dark:from-emerald-900/20 dark:to-teal-900/20 border border-emerald-200 dark:border-emerald-800 rounded-xl p-6">
        <div className="flex items-start gap-4">
          <div className="p-3 bg-emerald-100 dark:bg-emerald-800 rounded-lg">
            <Zap className="h-6 w-6 text-emerald-600 dark:text-emerald-400" />
          </div>
          <div className="flex-1">
            <h3 className="text-lg font-semibold text-emerald-900 dark:text-emerald-100 mb-2">
              Zero Code Required!
            </h3>
            <p className="text-sm text-emerald-700 dark:text-emerald-300 mb-4">
              Just change your database connection string to point to the sharding proxy.
              The proxy automatically routes queries to the correct shard based on the rules you configure here.
            </p>
            <div className="flex items-center gap-4">
              <div className="flex-1 bg-white dark:bg-gray-800 rounded-lg p-3 font-mono text-sm">
                <span className="text-gray-500">Before:</span> jdbc:postgresql://<span className="text-red-500">db-server</span>:5432/mydb<br />
                <span className="text-gray-500">After:</span>&nbsp; jdbc:postgresql://<span className="text-emerald-500">sharding-proxy</span>:5432/mydb
              </div>
              <Button variant="secondary" size="sm" onClick={copyConnectionString}>
                <Copy className="h-4 w-4 mr-1" />
                Copy
              </Button>
            </div>
          </div>
        </div>
      </div>

      {/* Search */}
      <div className="card">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-gray-400" />
          <input
            type="text"
            placeholder="Search databases..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary-500 focus:border-transparent"
          />
        </div>
      </div>

      {/* Databases List */}
      {filteredDatabases.length > 0 ? (
        <div className="space-y-6">
          {filteredDatabases.map(([dbKey, config]) => (
            <div key={dbKey} className="card">
              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-3">
                  <div className="p-2 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
                    <Database className="h-5 w-5 text-blue-600 dark:text-blue-400" />
                  </div>
                  <div>
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-white">{config.name}</h3>
                    <p className="text-sm text-gray-500 dark:text-gray-400">Database: <code>{dbKey}</code></p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-sm text-gray-500 dark:text-gray-400">
                    {config.sharding_rules?.length || 0} rules
                  </span>
                  <Button
                    size="sm"
                    variant="secondary"
                    onClick={() => {
                      setSelectedDatabase(dbKey);
                      setNewRule({ table: '', shard_key: '', strategy: 'hash', description: '' });
                      setShowAddRuleModal(true);
                    }}
                  >
                    <Plus className="h-4 w-4 mr-1" />
                    Add Rule
                  </Button>
                </div>
              </div>

              {/* Rules Table */}
              {config.sharding_rules && config.sharding_rules.length > 0 ? (
                <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                    <thead className="bg-gray-50 dark:bg-gray-800/50">
                      <tr>
                        <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                          Table
                        </th>
                        <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                          Shard Key
                        </th>
                        <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                          Strategy
                        </th>
                        <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                          Description
                        </th>
                        <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                          Actions
                        </th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                      {config.sharding_rules.map((rule) => (
                        <tr key={rule.table} className="hover:bg-gray-50 dark:hover:bg-gray-800/50">
                          <td className="px-4 py-3 whitespace-nowrap">
                            <div className="flex items-center gap-2">
                              <Table2 className="h-4 w-4 text-gray-400" />
                              <span className="font-medium text-gray-900 dark:text-white">{rule.table}</span>
                            </div>
                          </td>
                          <td className="px-4 py-3 whitespace-nowrap">
                            {rule.shard_key ? (
                              <div className="flex items-center gap-2">
                                <Key className="h-4 w-4 text-amber-500" />
                                <code className="px-2 py-1 bg-amber-50 dark:bg-amber-900/20 text-amber-700 dark:text-amber-300 rounded text-sm">
                                  {rule.shard_key}
                                </code>
                              </div>
                            ) : (
                              <span className="text-gray-400">-</span>
                            )}
                          </td>
                          <td className="px-4 py-3 whitespace-nowrap">
                            <StrategyBadge strategy={rule.strategy} />
                          </td>
                          <td className="px-4 py-3 text-sm text-gray-500 dark:text-gray-400 max-w-xs truncate">
                            {rule.description || '-'}
                          </td>
                          <td className="px-4 py-3 whitespace-nowrap text-right">
                            <div className="flex items-center justify-end gap-2">
                              <button
                                onClick={() => {
                                  setSelectedDatabase(dbKey);
                                  setEditingRule({ database: dbKey, rule });
                                  setNewRule(rule);
                                  setShowAddRuleModal(true);
                                }}
                                className="p-1 text-gray-400 hover:text-blue-600 transition-colors"
                                title="Edit rule"
                              >
                                <Edit2 className="h-4 w-4" />
                              </button>
                              <button
                                onClick={() => deleteRuleMutation.mutate({ database: dbKey, table: rule.table })}
                                className="p-1 text-gray-400 hover:text-red-600 transition-colors"
                                title="Delete rule"
                              >
                                <Trash2 className="h-4 w-4" />
                              </button>
                            </div>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              ) : (
                <div className="text-center py-8 text-gray-500 dark:text-gray-400">
                  <Table2 className="h-8 w-8 mx-auto mb-2 opacity-50" />
                  <p>No sharding rules configured for this database</p>
                  <Button
                    size="sm"
                    variant="secondary"
                    className="mt-3"
                    onClick={() => {
                      setSelectedDatabase(dbKey);
                      setNewRule({ table: '', shard_key: '', strategy: 'hash', description: '' });
                      setShowAddRuleModal(true);
                    }}
                  >
                    <Plus className="h-4 w-4 mr-1" />
                    Add First Rule
                  </Button>
                </div>
              )}
            </div>
          ))}
        </div>
      ) : (
        <div className="card">
          <div className="text-center py-12">
            <Database className="mx-auto h-12 w-12 text-gray-400" />
            <h3 className="mt-2 text-lg font-medium text-gray-900 dark:text-white">No databases configured</h3>
            <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
              Add a database to start configuring sharding rules
            </p>
            <Button className="mt-4" onClick={() => setShowAddDatabaseModal(true)}>
              <Plus className="h-4 w-4 mr-2" />
              Add Database
            </Button>
          </div>
        </div>
      )}

      {/* Add Database Modal */}
      <Modal
        isOpen={showAddDatabaseModal}
        onClose={() => setShowAddDatabaseModal(false)}
        title="Add Database"
        size="md"
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Application Name
            </label>
            <input
              type="text"
              value={newDatabase.name}
              onChange={(e) => setNewDatabase({ ...newDatabase, name: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
              placeholder="e.g., E-Commerce Application"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Database Name
            </label>
            <input
              type="text"
              value={newDatabase.database}
              onChange={(e) => setNewDatabase({ ...newDatabase, database: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
              placeholder="e.g., ecommerce_db"
            />
            <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
              This should match the database name your application connects to
            </p>
          </div>
        </div>
        <div className="flex gap-3 mt-6">
          <Button variant="secondary" className="flex-1" onClick={() => setShowAddDatabaseModal(false)}>
            Cancel
          </Button>
          <Button
            className="flex-1"
            disabled={!newDatabase.name || !newDatabase.database}
            isLoading={createRulesMutation.isPending}
            onClick={() => createRulesMutation.mutate({
              database: newDatabase.database,
              data: { name: newDatabase.name, sharding_rules: [] }
            })}
          >
            Add Database
          </Button>
        </div>
      </Modal>

      {/* Add/Edit Rule Modal */}
      <Modal
        isOpen={showAddRuleModal}
        onClose={() => {
          setShowAddRuleModal(false);
          setEditingRule(null);
          setNewRule({ table: '', shard_key: '', strategy: 'hash', description: '' });
        }}
        title={editingRule ? 'Edit Sharding Rule' : 'Add Sharding Rule'}
        size="md"
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Table Name
            </label>
            <input
              type="text"
              value={newRule.table}
              onChange={(e) => setNewRule({ ...newRule, table: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
              placeholder="e.g., users, orders"
              disabled={!!editingRule}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Strategy
            </label>
            <div className="grid grid-cols-3 gap-3">
              {[
                { value: 'hash', label: 'Hash', icon: Hash, description: 'Distribute by hash' },
                { value: 'range', label: 'Range', icon: ArrowRight, description: 'By value range' },
                { value: 'broadcast', label: 'Broadcast', icon: Radio, description: 'All shards' },
              ].map((strategy) => (
                <button
                  key={strategy.value}
                  type="button"
                  onClick={() => setNewRule({ ...newRule, strategy: strategy.value as ShardingRule['strategy'] })}
                  className={`p-3 border rounded-lg text-left transition-colors ${newRule.strategy === strategy.value
                      ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20'
                      : 'border-gray-200 dark:border-gray-700 hover:border-gray-300'
                    }`}
                >
                  <strategy.icon className={`h-5 w-5 mb-1 ${newRule.strategy === strategy.value ? 'text-primary-600' : 'text-gray-400'
                    }`} />
                  <div className="font-medium text-sm text-gray-900 dark:text-white">{strategy.label}</div>
                  <div className="text-xs text-gray-500">{strategy.description}</div>
                </button>
              ))}
            </div>
          </div>
          {newRule.strategy !== 'broadcast' && (
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Shard Key Column
              </label>
              <input
                type="text"
                value={newRule.shard_key}
                onChange={(e) => setNewRule({ ...newRule, shard_key: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                placeholder="e.g., user_id, customer_id"
              />
              <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                The column used to determine which shard stores the data
              </p>
            </div>
          )}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Description (optional)
            </label>
            <input
              type="text"
              value={newRule.description || ''}
              onChange={(e) => setNewRule({ ...newRule, description: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
              placeholder="e.g., Shard users by their ID"
            />
          </div>
        </div>
        <div className="flex gap-3 mt-6">
          <Button
            variant="secondary"
            className="flex-1"
            onClick={() => {
              setShowAddRuleModal(false);
              setEditingRule(null);
              setNewRule({ table: '', shard_key: '', strategy: 'hash', description: '' });
            }}
          >
            Cancel
          </Button>
          <Button
            className="flex-1"
            disabled={!newRule.table || (newRule.strategy !== 'broadcast' && !newRule.shard_key)}
            isLoading={addRuleMutation.isPending}
            onClick={() => {
              if (selectedDatabase) {
                addRuleMutation.mutate({ database: selectedDatabase, rule: newRule });
              }
            }}
          >
            {editingRule ? 'Update Rule' : 'Add Rule'}
          </Button>
        </div>
      </Modal>

      {/* Test Query Modal */}
      <Modal
        isOpen={showTestQueryModal}
        onClose={() => {
          setShowTestQueryModal(false);
          setTestResult(null);
          setTestQueryInput({ database: '', query: '' });
        }}
        title="Test Query Routing"
        size="lg"
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Database
            </label>
            <select
              value={testQueryInput.database}
              onChange={(e) => setTestQueryInput({ ...testQueryInput, database: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
            >
              <option value="">Select database...</option>
              {databases.map(([key]) => (
                <option key={key} value={key}>{key}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              SQL Query
            </label>
            <textarea
              value={testQueryInput.query}
              onChange={(e) => setTestQueryInput({ ...testQueryInput, query: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white font-mono text-sm"
              rows={4}
              placeholder="SELECT * FROM users WHERE user_id = '123'"
            />
          </div>
          <Button
            className="w-full"
            disabled={!testQueryInput.database || !testQueryInput.query}
            isLoading={testQueryMutation.isPending}
            onClick={() => testQueryMutation.mutate(testQueryInput)}
          >
            <TestTube2 className="h-4 w-4 mr-2" />
            Test Routing
          </Button>

          {testResult && (
            <div className="mt-4 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
              <h4 className="font-medium text-gray-900 dark:text-white mb-3 flex items-center gap-2">
                <Code2 className="h-4 w-4" />
                Routing Result
              </h4>
              <div className="space-y-2 text-sm">
                <div className="flex items-center gap-2">
                  <span className="text-gray-500 w-24">Table:</span>
                  <code className="px-2 py-1 bg-gray-200 dark:bg-gray-700 rounded">{testResult.table || 'N/A'}</code>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-gray-500 w-24">Shard Key:</span>
                  <code className="px-2 py-1 bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-300 rounded">
                    {testResult.shard_key || 'N/A'}
                  </code>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-gray-500 w-24">Value:</span>
                  <code className="px-2 py-1 bg-gray-200 dark:bg-gray-700 rounded">{testResult.shard_value || 'N/A'}</code>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-gray-500 w-24">Routing:</span>
                  {testResult.routing === 'single_shard' ? (
                    <span className="flex items-center gap-1 text-green-600 dark:text-green-400">
                      <CheckCircle2 className="h-4 w-4" />
                      Single Shard → {testResult.target_shard}
                    </span>
                  ) : testResult.routing === 'scatter_gather' ? (
                    <span className="flex items-center gap-1 text-amber-600 dark:text-amber-400">
                      <Radio className="h-4 w-4" />
                      Scatter-Gather (all shards)
                    </span>
                  ) : (
                    <span className="text-gray-600 dark:text-gray-400">{testResult.routing}</span>
                  )}
                </div>
                {testResult.target_endpoint && (
                  <div className="flex items-start gap-2">
                    <span className="text-gray-500 w-24">Endpoint:</span>
                    <code className="px-2 py-1 bg-gray-200 dark:bg-gray-700 rounded text-xs break-all">
                      {testResult.target_endpoint}
                    </code>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      </Modal>
    </div>
  );
}

// Strategy Badge Component
function StrategyBadge({ strategy }: { strategy: string }) {
  const config = {
    hash: { icon: Hash, color: 'blue', label: 'Hash' },
    range: { icon: ArrowRight, color: 'purple', label: 'Range' },
    broadcast: { icon: Radio, color: 'green', label: 'Broadcast' },
  }[strategy] || { icon: Hash, color: 'gray', label: strategy };

  const Icon = config.icon;

  return (
    <span className={`inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium
      ${config.color === 'blue' ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300' : ''}
      ${config.color === 'purple' ? 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-300' : ''}
      ${config.color === 'green' ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300' : ''}
      ${config.color === 'gray' ? 'bg-gray-100 text-gray-700 dark:bg-gray-900/30 dark:text-gray-300' : ''}
    `}>
      <Icon className="h-3 w-3" />
      {config.label}
    </span>
  );
}

