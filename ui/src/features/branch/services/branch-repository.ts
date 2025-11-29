import { ApiFactory } from '@/core/http/api-factory';
import type { Branch, CreateBranchRequest } from '../types';

export class BranchRepository {
  private client = ApiFactory.getManagerClient();

  async listBranches(dbName: string): Promise<Branch[]> {
    return this.client.get<Branch[]>(`/databases/${dbName}/branches`);
  }

  async createBranch(dbName: string, request: CreateBranchRequest): Promise<Branch> {
    return this.client.post<Branch>(`/databases/${dbName}/branches`, request);
  }

  async getBranch(branchID: string): Promise<Branch> {
    return this.client.get<Branch>(`/branches/${branchID}`);
  }

  async deleteBranch(branchID: string): Promise<void> {
    await this.client.delete(`/branches/${branchID}`);
  }

  async mergeBranch(branchID: string): Promise<void> {
    await this.client.post(`/branches/${branchID}/merge`);
  }
}

export const branchRepository = new BranchRepository();

