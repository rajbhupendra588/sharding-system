import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 20 }, // Ramp up to 20 users
    { duration: '1m', target: 20 },  // Stay at 20 users
    { duration: '30s', target: 0 },  // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests should be below 500ms
  },
};

const BASE_URL = 'http://localhost:8080/api/v1';

export default function () {
  // Simulate a read operation (GET /shards)
  // Note: In a real scenario, we would query data, but for now we check shard health/metadata
  // assuming the router forwards this correctly or we hit a specific endpoint.
  // Adjust endpoint as per actual API.
  
  // Let's try to get a specific key to test routing
  const key = `user-${Math.floor(Math.random() * 1000)}`;
  
  // We'll assume a hypothetical endpoint that routes based on key
  // If not available, we'll hit the health endpoint to test basic throughput
  const res = http.get(`${BASE_URL}/health`);

  check(res, {
    'status is 200': (r) => r.status === 200,
  });

  sleep(1);
}
