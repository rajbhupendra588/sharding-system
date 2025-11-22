import { useState } from 'react';
import { useMutation } from '@tanstack/react-query';
import { Play, Copy, CheckCircle } from 'lucide-react';
import { apiClient } from '@/lib/api-client';
import { toast } from 'react-hot-toast';
import Button from '@/components/ui/Button';
import Input from '@/components/ui/Input';
import StatusBadge from '@/components/ui/StatusBadge';
import { formatDuration } from '@/lib/utils';
import type { QueryRequest, QueryResponse } from '@/types';

export default function QueryExecutor() {
  const [shardKey, setShardKey] = useState('');
  const [query, setQuery] = useState('');
  const [params, setParams] = useState('');
  const [consistency, setConsistency] = useState<'strong' | 'eventual'>('strong');
  const [result, setResult] = useState<QueryResponse | null>(null);
  const [copied, setCopied] = useState(false);

  const executeMutation = useMutation({
    mutationFn: (request: QueryRequest) => apiClient.executeQuery(request),
    onSuccess: (data) => {
      setResult(data);
      toast.success('Query executed successfully');
    },
    onError: (error: { message: string }) => {
      toast.error(`Query failed: ${error.message}`);
      setResult(null);
    },
  });

  const handleExecute = () => {
    if (!shardKey.trim() || !query.trim()) {
      toast.error('Shard key and query are required');
      return;
    }

    let parsedParams: unknown[] = [];
    if (params.trim()) {
      try {
        parsedParams = JSON.parse(`[${params}]`);
      } catch {
        toast.error('Invalid parameters format. Use comma-separated values.');
        return;
      }
    }

    executeMutation.mutate({
      shard_key: shardKey,
      query: query.trim(),
      params: parsedParams,
      consistency,
    });
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">Query Executor</h1>
        <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
          Execute SQL queries against sharded databases
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Query Input */}
        <div className="card space-y-4">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">Query Configuration</h2>

          <Input
            label="Shard Key"
            value={shardKey}
            onChange={(e) => setShardKey(e.target.value)}
            placeholder="user-123"
            required
          />

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              SQL Query
            </label>
            <textarea
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent font-mono text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-white"
              rows={8}
              placeholder="SELECT * FROM users WHERE id = $1"
            />
          </div>

          <Input
            label="Parameters (comma-separated)"
            value={params}
            onChange={(e) => setParams(e.target.value)}
            placeholder="user-123, 100"
            helperText="Use comma-separated values. Example: user-123, 100"
          />

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Consistency Level
            </label>
            <select
              value={consistency}
              onChange={(e) => setConsistency(e.target.value as 'strong' | 'eventual')}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent bg-white dark:bg-gray-800 text-gray-900 dark:text-white"
            >
              <option value="strong">Strong (Read from primary)</option>
              <option value="eventual">Eventual (Can read from replica)</option>
            </select>
          </div>

          <Button
            onClick={handleExecute}
            isLoading={executeMutation.isPending}
            className="w-full"
          >
            <Play className="h-4 w-4 mr-2" />
            Execute Query
          </Button>
        </div>

        {/* Results */}
        <div className="card space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-gray-900 dark:text-white">Results</h2>
            {result && (
              <button
                onClick={() => copyToClipboard(JSON.stringify(result, null, 2))}
                className="text-sm text-primary-600 hover:text-primary-700 flex items-center"
              >
                {copied ? (
                  <>
                    <CheckCircle className="h-4 w-4 mr-1" />
                    Copied
                  </>
                ) : (
                  <>
                    <Copy className="h-4 w-4 mr-1" />
                    Copy JSON
                  </>
                )}
              </button>
            )}
          </div>

          {executeMutation.isPending ? (
            <div className="flex items-center justify-center py-12">
              <div className="text-center">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600 mx-auto mb-2"></div>
                <p className="text-sm text-gray-500 dark:text-gray-400">Executing query...</p>
              </div>
            </div>
          ) : result ? (
            <div className="space-y-4">
              {/* Summary */}
              <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4 space-y-2">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Shard ID:</span>
                  <StatusBadge status={result.shard_id} />
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Row Count:</span>
                  <span className="text-sm text-gray-900 dark:text-white">{result.row_count}</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Latency:</span>
                  <span className="text-sm text-gray-900 dark:text-white">{formatDuration(result.latency_ms)}</span>
                </div>
              </div>

              {/* Results Table */}
              {result.rows.length > 0 ? (
                <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                    <thead className="bg-gray-50 dark:bg-gray-800">
                      <tr>
                        {Object.keys(result.rows[0]).map((key) => (
                          <th
                            key={key}
                            className="px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase"
                          >
                            {key}
                          </th>
                        ))}
                      </tr>
                    </thead>
                    <tbody className="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-700">
                      {result.rows.map((row, index) => (
                        <tr key={index}>
                          {Object.entries(row).map(([key, value]) => (
                            <td
                              key={key}
                              className="px-4 py-2 text-sm text-gray-900 dark:text-gray-300 whitespace-nowrap"
                            >
                              {typeof value === 'object' ? JSON.stringify(value) : String(value)}
                            </td>
                          ))}
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              ) : (
                <div className="text-center py-8 text-gray-500">
                  No rows returned
                </div>
              )}
            </div>
          ) : (
            <div className="text-center py-12 text-gray-500 dark:text-gray-400">
              <p>No query executed yet</p>
              <p className="text-sm mt-1">Enter a query and click Execute to see results</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

