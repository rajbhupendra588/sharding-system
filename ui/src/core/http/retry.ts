/**
 * Retry Logic
 * Exponential backoff retry strategy for failed requests
 */

export interface RetryConfig {
  maxAttempts: number;
  baseDelay: number;
  maxDelay: number;
  factor: number;
  retryableStatuses: number[];
}

const DEFAULT_RETRY_CONFIG: RetryConfig = {
  maxAttempts: 3,
  baseDelay: 1000, // 1 second
  maxDelay: 10000, // 10 seconds
  factor: 2,
  retryableStatuses: [408, 429, 500, 502, 503, 504],
};

export function isRetryableError(error: { code?: string; status?: number }): boolean {
  // Network errors are retryable
  if (error.code === 'NETWORK_ERROR' || error.code === 'ECONNABORTED') {
    return true;
  }

  // Check if status code is retryable
  if (error.status && DEFAULT_RETRY_CONFIG.retryableStatuses.includes(error.status)) {
    return true;
  }

  return false;
}

export function calculateDelay(attempt: number, config: RetryConfig = DEFAULT_RETRY_CONFIG): number {
  const delay = config.baseDelay * Math.pow(config.factor, attempt - 1);
  return Math.min(delay, config.maxDelay);
}

export async function retryWithBackoff<T>(
  fn: () => Promise<T>,
  config: RetryConfig = DEFAULT_RETRY_CONFIG
): Promise<T> {
  let lastError: unknown;

  for (let attempt = 1; attempt <= config.maxAttempts; attempt++) {
    try {
      return await fn();
    } catch (error) {
      lastError = error;

      // Don't retry if error is not retryable
      if (!isRetryableError(error as { code?: string; status?: number })) {
        throw error;
      }

      // Don't retry on last attempt
      if (attempt === config.maxAttempts) {
        throw error;
      }

      // Wait before retrying
      const delay = calculateDelay(attempt, config);
      await new Promise((resolve) => setTimeout(resolve, delay));
    }
  }

  throw lastError;
}

