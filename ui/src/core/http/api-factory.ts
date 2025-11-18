/**
 * API Factory
 * Creates HTTP clients for different services
 */

import { HttpClient } from './client';
import { API_CONFIG, appConfig } from '../config';

export interface HttpClientConfig {
  baseURL: string;
  timeout?: number;
  headers?: Record<string, string>;
  serviceName?: string;
}

export class ApiFactory {
  private static managerClient: HttpClient;
  private static routerClient: HttpClient;

  static getManagerClient(): HttpClient {
    if (!this.managerClient) {
      const config = appConfig.getConfig();
      this.managerClient = new HttpClient({
        baseURL: `${config.managerUrl}${API_CONFIG.MANAGER_API_PREFIX}`,
        serviceName: 'manager',
      });
    }
    return this.managerClient;
  }

  static getRouterClient(): HttpClient {
    if (!this.routerClient) {
      const config = appConfig.getConfig();
      this.routerClient = new HttpClient({
        baseURL: `${config.routerUrl}${API_CONFIG.ROUTER_API_PREFIX}`,
        serviceName: 'router',
      });
    }
    return this.routerClient;
  }

  static resetClients(): void {
    this.managerClient = null as unknown as HttpClient;
    this.routerClient = null as unknown as HttpClient;
  }
}

