/**
 * Application Configuration
 * Runtime configuration management
 */

import { API_CONFIG, STORAGE_KEYS } from './constants';

export interface AppConfig {
  managerUrl: string;
  routerUrl: string;
  refreshInterval: number;
}

class AppConfigManager {
  private config: AppConfig;

  constructor() {
    this.config = this.loadConfig();
  }

  private loadConfig(): AppConfig {
    // In K8s, use relative URLs - nginx will proxy to services
    const managerUrl = localStorage.getItem(STORAGE_KEYS.MANAGER_URL) || API_CONFIG.MANAGER_BASE_URL || '';
    const routerUrl = localStorage.getItem(STORAGE_KEYS.ROUTER_URL) || API_CONFIG.ROUTER_BASE_URL || '';
    return {
      managerUrl: managerUrl || '', // Empty string = relative URL
      routerUrl: routerUrl || '', // Empty string = relative URL
      refreshInterval: parseInt(
        localStorage.getItem(STORAGE_KEYS.REFRESH_INTERVAL) || '10000',
        10
      ),
    };
  }

  getConfig(): AppConfig {
    return { ...this.config };
  }

  updateConfig(updates: Partial<AppConfig>): void {
    this.config = { ...this.config, ...updates };
    this.saveConfig();
  }

  private saveConfig(): void {
    localStorage.setItem(STORAGE_KEYS.MANAGER_URL, this.config.managerUrl);
    localStorage.setItem(STORAGE_KEYS.ROUTER_URL, this.config.routerUrl);
    localStorage.setItem(STORAGE_KEYS.REFRESH_INTERVAL, this.config.refreshInterval.toString());
  }

  reset(): void {
    localStorage.removeItem(STORAGE_KEYS.MANAGER_URL);
    localStorage.removeItem(STORAGE_KEYS.ROUTER_URL);
    localStorage.removeItem(STORAGE_KEYS.REFRESH_INTERVAL);
    this.config = this.loadConfig();
  }
}

export const appConfig = new AppConfigManager();

