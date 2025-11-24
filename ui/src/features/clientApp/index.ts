import { useQuery } from '@tanstack/react-query';

export interface ClientApp {
  id: string;
  name: string;
  description?: string;
  database_name?: string; // Database name for which sharding needs to be created
  database_host?: string; // Database host
  database_port?: string; // Database port
  database_user?: string; // Database user
  database_password?: string; // Database password
  key_prefix?: string;
  namespace?: string;
  cluster_name?: string;
  status: 'active' | 'inactive';
  created_at: string;
  updated_at: string;
  last_seen: string;
  shard_ids: string[];
  request_count: number;
}

const API_BASE = '/api/v1';

export function useClientApps() {
  return useQuery<ClientApp[]>({
    queryKey: ['clientApps'],
    queryFn: async () => {
      const response = await fetch(`${API_BASE}/client-apps`);
      if (!response.ok) {
        throw new Error('Failed to fetch client applications');
      }
      return response.json();
    },
  });
}

export function useClientApp(id: string) {
  return useQuery<ClientApp>({
    queryKey: ['clientApp', id],
    queryFn: async () => {
      const response = await fetch(`${API_BASE}/client-apps/${id}`);
      if (!response.ok) {
        throw new Error('Failed to fetch client application');
      }
      return response.json();
    },
    enabled: !!id,
  });
}

