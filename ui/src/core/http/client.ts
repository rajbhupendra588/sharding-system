/**
 * HTTP Client
 * Base HTTP client with interceptors and error handling
 */

import axios, { AxiosInstance, AxiosError, InternalAxiosRequestConfig, AxiosResponse } from 'axios';
import { API_CONFIG, STORAGE_KEYS } from '../config/constants';
import { ApiError } from '@/shared/types';
import { retryWithBackoff } from './retry';
import { getCircuitBreaker } from './circuit-breaker';

export interface HttpClientConfig {
  baseURL: string;
  timeout?: number;
  headers?: Record<string, string>;
  serviceName?: string;
}

export class HttpClient {
  private client: AxiosInstance;
  private serviceName: string;

  constructor(config: HttpClientConfig & { serviceName?: string }) {
    this.serviceName = config.serviceName || 'default';
    this.client = axios.create({
      baseURL: config.baseURL,
      timeout: config.timeout || API_CONFIG.TIMEOUT,
      headers: {
        'Content-Type': 'application/json',
        ...config.headers,
      },
    });

    this.setupInterceptors();
  }

  private setupInterceptors(): void {
    // Request interceptor
    this.client.interceptors.request.use(
      (config: InternalAxiosRequestConfig) => {
        const token = localStorage.getItem(STORAGE_KEYS.AUTH_TOKEN);
        if (token && config.headers) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => Promise.reject(error)
    );

    // Response interceptor
    this.client.interceptors.response.use(
      (response: AxiosResponse) => response,
      (error: AxiosError) => {
        return Promise.reject(this.handleError(error));
      }
    );
  }

  private handleError(error: AxiosError): ApiError {
    if (error.response) {
      // Server responded with error status
      let message: string;

      // Handle both JSON and plain text error responses
      if (typeof error.response.data === 'string') {
        message = error.response.data;
      } else if (error.response.data && typeof error.response.data === 'object') {
        message = (error.response.data as { message?: string; error?: { message?: string } })?.message
          || (error.response.data as { error?: { message?: string } })?.error?.message
          || error.message;
      } else {
        message = error.message;
      }

      return {
        message,
        code: error.response.status.toString(),
        details: error.response.data,
      };
    }

    if (error.request) {
      // Request made but no response received
      return {
        message: 'Network error: Unable to reach server',
        code: 'NETWORK_ERROR',
      };
    }

    // Something else happened
    return {
      message: error.message,
      code: 'UNKNOWN_ERROR',
    };
  }

  async get<T>(url: string, config?: InternalAxiosRequestConfig): Promise<T> {
    const circuitBreaker = getCircuitBreaker(this.serviceName);
    return circuitBreaker.execute(() =>
      retryWithBackoff(async () => {
        const response = await this.client.get<T>(url, config);
        return response.data;
      })
    );
  }

  async post<T>(url: string, data?: unknown, config?: InternalAxiosRequestConfig): Promise<T> {
    const circuitBreaker = getCircuitBreaker(this.serviceName);
    return circuitBreaker.execute(() =>
      retryWithBackoff(async () => {
        const response = await this.client.post<T>(url, data, config);
        return response.data;
      })
    );
  }

  async put<T>(url: string, data?: unknown, config?: InternalAxiosRequestConfig): Promise<T> {
    const circuitBreaker = getCircuitBreaker(this.serviceName);
    return circuitBreaker.execute(() =>
      retryWithBackoff(async () => {
        const response = await this.client.put<T>(url, data, config);
        return response.data;
      })
    );
  }

  async patch<T>(url: string, data?: unknown, config?: InternalAxiosRequestConfig): Promise<T> {
    const circuitBreaker = getCircuitBreaker(this.serviceName);
    return circuitBreaker.execute(() =>
      retryWithBackoff(async () => {
        const response = await this.client.patch<T>(url, data, config);
        return response.data;
      })
    );
  }

  async delete<T>(url: string, config?: InternalAxiosRequestConfig): Promise<T> {
    const circuitBreaker = getCircuitBreaker(this.serviceName);
    return circuitBreaker.execute(() =>
      retryWithBackoff(async () => {
        const response = await this.client.delete<T>(url, config);
        return response.data;
      })
    );
  }
}

