import { describe, it, expect, vi, beforeEach } from 'vitest';
import axios, { AxiosError } from 'axios';
import { HttpClient } from './client';
import { getCircuitBreaker } from './circuit-breaker';
import { retryWithBackoff } from './retry';

vi.mock('axios');
vi.mock('./circuit-breaker');
vi.mock('./retry');

describe('HttpClient', () => {
  let client: HttpClient;
  const mockAxiosInstance = {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
    interceptors: {
      request: { use: vi.fn() },
      response: { use: vi.fn() },
    },
  };

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(axios.create).mockReturnValue(mockAxiosInstance as any);
    vi.mocked(getCircuitBreaker).mockReturnValue({
      execute: vi.fn((fn) => fn()),
    } as any);
    vi.mocked(retryWithBackoff).mockImplementation(async (fn) => fn());
    
    client = new HttpClient({
      baseURL: 'http://localhost:8080',
      serviceName: 'test-service',
    });
  });

  describe('get', () => {
    it('should make GET request', async () => {
      const mockResponse = { data: { id: '1', name: 'test' } };
      mockAxiosInstance.get.mockResolvedValue(mockResponse);

      const result = await client.get('/api/test');

      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/api/test', undefined);
      expect(result).toEqual(mockResponse.data);
    });

    it('should handle errors', async () => {
      const mockError = new Error('Network error') as AxiosError;
      mockAxiosInstance.get.mockRejectedValue(mockError);

      await expect(client.get('/api/test')).rejects.toThrow();
    });
  });

  describe('post', () => {
    it('should make POST request', async () => {
      const mockData = { name: 'test' };
      const mockResponse = { data: { id: '1', ...mockData } };
      mockAxiosInstance.post.mockResolvedValue(mockResponse);

      const result = await client.post('/api/test', mockData);

      expect(mockAxiosInstance.post).toHaveBeenCalledWith('/api/test', mockData, undefined);
      expect(result).toEqual(mockResponse.data);
    });
  });

  describe('error handling', () => {
    it('should handle network errors', () => {
      const error = {
        request: {},
        message: 'Network error',
      } as AxiosError;

      // Test error handling through interceptor
      const responseInterceptor = mockAxiosInstance.interceptors.response.use.mock.calls[0][1];
      const handledError = responseInterceptor(error);

      expect(handledError).toBeDefined();
    });
  });
});

