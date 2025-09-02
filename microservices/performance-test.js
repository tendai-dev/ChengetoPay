import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');

export const options = {
  stages: [
    { duration: '30s', target: 10 },   // Ramp up to 10 users
    { duration: '1m', target: 50 },    // Stay at 50 users
    { duration: '2m', target: 100 },   // Ramp up to 100 users
    { duration: '2m', target: 200 },   // Peak load at 200 users
    { duration: '1m', target: 100 },   // Scale down
    { duration: '30s', target: 0 },    // Ramp down to 0
  ],
  thresholds: {
    'http_req_duration': ['p(95)<500'], // 95% of requests must complete below 500ms
    'errors': ['rate<0.1'],              // Error rate must be below 10%
  },
};

const BASE_URL = 'http://localhost:8090';

export default function () {
  // Test 1: Health check
  let healthRes = http.get(`${BASE_URL}/health`);
  check(healthRes, {
    'health check status is 200': (r) => r.status === 200,
  });
  errorRate.add(healthRes.status !== 200);

  sleep(1);

  // Test 2: Create escrow
  const escrowPayload = JSON.stringify({
    amount: Math.random() * 1000,
    currency: 'USD',
    buyer_id: 'buyer_' + Math.random(),
    seller_id: 'seller_' + Math.random(),
  });

  const escrowRes = http.post(`${BASE_URL}/api/v1/escrows`, escrowPayload, {
    headers: { 'Content-Type': 'application/json' },
  });

  check(escrowRes, {
    'escrow creation status is 201': (r) => r.status === 201,
    'escrow has ID': (r) => JSON.parse(r.body).id !== undefined,
  });
  errorRate.add(escrowRes.status !== 201);

  sleep(2);

  // Test 3: Process payment
  const paymentPayload = JSON.stringify({
    amount: Math.random() * 500,
    currency: 'USD',
    payment_method: 'card',
  });

  const paymentRes = http.post(`${BASE_URL}/api/v1/payments`, paymentPayload, {
    headers: { 'Content-Type': 'application/json' },
  });

  check(paymentRes, {
    'payment status is 200': (r) => r.status === 200,
  });
  errorRate.add(paymentRes.status !== 200);

  sleep(1);

  // Test 4: Check service status
  const services = ['escrow', 'payment', 'ledger', 'risk'];
  const serviceIndex = Math.floor(Math.random() * services.length);
  
  const statusRes = http.get(`${BASE_URL}/api/v1/${services[serviceIndex]}/status`);
  check(statusRes, {
    'service status is 200': (r) => r.status === 200,
  });
  errorRate.add(statusRes.status !== 200);

  sleep(1);
}

export function handleSummary(data) {
  return {
    'performance-report.html': htmlReport(data),
    'performance-summary.json': JSON.stringify(data),
  };
}

function htmlReport(data) {
  return `
<!DOCTYPE html>
<html>
<head>
    <title>Performance Test Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        h1 { color: #333; }
        .metric { margin: 20px 0; padding: 15px; background: #f5f5f5; border-radius: 5px; }
        .pass { color: green; }
        .fail { color: red; }
    </style>
</head>
<body>
    <h1>Performance Test Results</h1>
    <div class="metric">
        <h3>Request Duration (95th percentile)</h3>
        <p>${data.metrics.http_req_duration.values['p(95)']}ms</p>
    </div>
    <div class="metric">
        <h3>Total Requests</h3>
        <p>${data.metrics.http_reqs.values.count}</p>
    </div>
    <div class="metric">
        <h3>Error Rate</h3>
        <p>${data.metrics.errors.values.rate * 100}%</p>
    </div>
</body>
</html>
  `;
}
