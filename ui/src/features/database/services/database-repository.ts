import { ApiFactory } from '@/core/http/api-factory';
import type {
  Database,
  DatabaseTemplate,
  CreateDatabaseRequest,
  DatabaseStatus,
  Backup,
  FailoverStatus,
  FailoverEvent,
} from '../types';

export class DatabaseRepository {
  private client = ApiFactory.getManagerClient();

  // Templates
  async getTemplates(): Promise<DatabaseTemplate[]> {
    return this.client.get<DatabaseTemplate[]>('/databases/templates');
  }

  // Database operations
  async create(request: CreateDatabaseRequest): Promise<Database> {
    return this.client.post<Database>('/databases', request);
  }

  async findAll(): Promise<Database[]> {
    return this.client.get<Database[]>('/databases');
  }

  async findById(id: string): Promise<Database> {
    return this.client.get<Database>(`/databases/${id}`);
  }

  async getStatus(id: string): Promise<DatabaseStatus> {
    return this.client.get<DatabaseStatus>(`/databases/${id}/status`);
  }

  // Backup operations
  async createBackup(databaseId: string, type: 'full' | 'incremental' = 'full'): Promise<Backup> {
    return this.client.post<Backup>(`/databases/${databaseId}/backups`, { type });
  }

  async listBackups(databaseId: string): Promise<Backup[]> {
    return this.client.get<Backup[]>(`/databases/${databaseId}/backups`);
  }

  async getBackup(databaseId: string, backupId: string): Promise<Backup> {
    return this.client.get<Backup>(`/databases/${databaseId}/backups/${backupId}`);
  }

  async restoreBackup(databaseId: string, backupId: string, targetDatabaseId?: string): Promise<void> {
    await this.client.post(`/databases/${databaseId}/backups/${backupId}/restore`, {
      target_database_id: targetDatabaseId || databaseId,
    });
  }

  async scheduleBackup(databaseId: string, schedule: string): Promise<void> {
    await this.client.post(`/databases/${databaseId}/backups/schedule`, { schedule });
  }

  // Failover operations
  async getFailoverStatus(): Promise<FailoverStatus> {
    return this.client.get<FailoverStatus>('/failover/status');
  }

  async enableFailover(): Promise<void> {
    await this.client.post('/failover/enable');
  }

  async disableFailover(): Promise<void> {
    await this.client.post('/failover/disable');
  }

  async getFailoverHistory(shardId?: string): Promise<FailoverEvent[]> {
    const url = shardId
      ? `/failover/history?shard_id=${shardId}`
      : '/failover/history';
    return this.client.get<FailoverEvent[]>(url);
  }
}

export const databaseRepository = new DatabaseRepository();

