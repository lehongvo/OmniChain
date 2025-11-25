import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const orderLatency = new Trend('order_latency');

export const options = {
  stages: [
    { duration: '5m', target: 10000 }, // Ramp-up: 0 → 10,000 users over 5 min
    { duration: '30m', target: 10000 }, // Sustained: 10,000 users for 30 min
    { duration: '5m', target: 50000 }, // Spike: 50,000 users for 5 min
    { duration: '5m', target: 0 }, // Ramp-down
  ],
  thresholds: {
    http_req_duration: ['p(95)<200'], // 95% of requests must be below 200ms
    http_req_failed: ['rate<0.01'], // Error rate must be less than 1%
    errors: ['rate<0.01'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const ACCESS_TOKEN = __ENV.ACCESS_TOKEN || '';

export default function () {
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${ACCESS_TOKEN}`,
  };

  // Create order
  const orderPayload = JSON.stringify({
    store_id: 'store-123',
    items: [
      {
        product_id: 'product-1',
        name: 'Test Product',
        quantity: 2,
        price: 10.99,
        subtotal: 21.98,
      },
    ],
  });

  const createStart = Date.now();
  const createRes = http.post(`${BASE_URL}/api/v1/orders`, orderPayload, { headers });
  const createLatency = Date.now() - createStart;
  
  orderLatency.add(createLatency);
  
  const createSuccess = check(createRes, {
    'order created status 201': (r) => r.status === 201,
    'order created response time < 200ms': (r) => r.timings.duration < 200,
  });

  if (!createSuccess) {
    errorRate.add(1);
  }

  sleep(1);

  // Get orders
  const getRes = http.get(`${BASE_URL}/api/v1/orders?page=1&page_size=20`, { headers });
  
  check(getRes, {
    'get orders status 200': (r) => r.status === 200,
    'get orders response time < 200ms': (r) => r.timings.duration < 200,
  });

  sleep(1);
}

