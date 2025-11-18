/**
 * Request Deduplication
 * Prevents duplicate concurrent requests
 */

interface PendingRequest<T> {
  promise: Promise<T>;
  timestamp: number;
}

const pendingRequests = new Map<string, PendingRequest<unknown>>();
const REQUEST_TIMEOUT = 60000; // 1 minute

export function deduplicateRequest<T>(
  key: string,
  requestFn: () => Promise<T>
): Promise<T> {
  // Check if there's a pending request
  const pending = pendingRequests.get(key);
  if (pending) {
    // Check if request is still valid (not timed out)
    if (Date.now() - pending.timestamp < REQUEST_TIMEOUT) {
      return pending.promise as Promise<T>;
    }
    // Request timed out, remove it
    pendingRequests.delete(key);
  }

  // Create new request
  const promise = requestFn().finally(() => {
    // Clean up after request completes
    pendingRequests.delete(key);
  });

  pendingRequests.set(key, {
    promise,
    timestamp: Date.now(),
  });

  return promise;
}

export function clearPendingRequests(): void {
  pendingRequests.clear();
}

export function getPendingRequestCount(): number {
  return pendingRequests.size;
}

