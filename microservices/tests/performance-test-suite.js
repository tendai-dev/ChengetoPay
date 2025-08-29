import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const responseTime = new Trend('response_time');

// Test configuration
export const options = {
  stages: [
    { duration: '2m', target: 10 },  // Ramp up to 10 users
    { duration: '5m', target: 10 },  // Stay at 10 users
    { duration: '2m', target: 50 },  // Ramp up to 50 users
    { duration: '5m', target: 50 },  // Stay at 50 users
    { duration: '2m', target: 100 }, // Ramp up to 100 users
    { duration: '5m', target: 100 }, // Stay at 100 users
    { duration: '2m', target: 0 },   // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests must complete below 500ms
    http_req_failed: ['rate<0.1'],    // Error rate must be less than 10%
    errors: ['rate<0.1'],             // Custom error rate must be less than 10%
  },
};

// Service configurations
const services = {
  'api-gateway': { port: 8090, endpoints: ['/', '/health', '/api/v1/health'] },
  'database-service': { port: 8115, endpoints: ['/health', '/v1/status', '/v1/health'] },
  'monitoring-service': { port: 8116, endpoints: ['/health', '/v1/health', '/v1/metrics'] },
  'message-queue-service': { port: 8117, endpoints: ['/health', '/v1/status', '/v1/events'] },
  'service-discovery': { port: 8118, endpoints: ['/health', '/v1/services', '/v1/health'] },
  'vault-service': { port: 8119, endpoints: ['/health', '/v1/status', '/v1/secrets'] },
  'escrow-service': { port: 8081, endpoints: ['/health', '/v1/info', '/v1/version'] },
  'payment-service': { port: 8083, endpoints: ['/health', '/v1/info', '/v1/version'] },
  'ledger-service': { port: 8084, endpoints: ['/health', '/v1/info', '/v1/version'] },
  'risk-service': { port: 8085, endpoints: ['/health', '/v1/info', '/v1/version'] },
  'treasury-service': { port: 8086, endpoints: ['/health', '/v1/info', '/v1/version'] },
  'evidence-service': { port: 8087, endpoints: ['/health', '/v1/info', '/v1/version'] },
  'compliance-service': { port: 8088, endpoints: ['/health', '/v1/info', '/v1/version'] },
  'workflow-service': { port: 8089, endpoints: ['/health', '/v1/info', '/v1/version'] },
  'journal-service': { port: 8091, endpoints: ['/health', '/v1/info', '/v1/version'] },
  'fees-service': { port: 8092, endpoints: ['/health', '/v1/info', '/v1/version'] },
  'refunds-service': { port: 8093, endpoints: ['/health', '/v1/info', '/v1/version'] },
  'transfers-service': { port: 8094, endpoints: ['/health', '/v1/info', '/v1/version'] },
  'fx-service': { port: 8095, endpoints: ['/health', '/v1/info', '/v1/version'] },
  'payouts-service': { port: 8096, endpoints: ['/health', '/v1/info', '/v1/version'] },
  'reserves-service': { port: 8097, endpoints: ['/health', '/v1/info', '/v1/version'] },
  'reconciliation-service': { port: 8098, endpoints: ['/health', '/v1/info', '/v1/version'] },
  'kyb-service': { port: 8099, endpoints: ['/health', '/v1/info', '/v1/version'] },
  'sca-service': { port: 8100, endpoints: ['/health', '/v1/info', '/v1/version'] },
  'disputes-service': { port: 8101, endpoints: ['/health', '/v1/info', '/v1/version'] },
  'dx-service': { port: 8102, endpoints: ['/health', '/v1/info', '/v1/version'] },
  'auth-service': { port: 8103, endpoints: ['/health', '/v1/info', '/v1/version'] },
  'idempotency-service': { port: 8104, endpoints: ['/health', '/v1/info', '/v1/version'] },
  'eventbus-service': { port: 8105, endpoints: ['/health', '/v1/info', '/v1/version'] },
  'saga-service': { port: 8106, endpoints: ['/health', '/v1/info', '/v1/version'] },
  'webhooks-service': { port: 8108, endpoints: ['/health', '/v1/info', '/v1/version'] },
  'observability-service': { port: 8109, endpoints: ['/health', '/v1/info', '/v1/version'] },
  'config-service': { port: 8110, endpoints: ['/health', '/v1/info', '/v1/version'] },
  'workers-service': { port: 8111, endpoints: ['/health', '/v1/info', '/v1/version'] },
  'portal-service': { port: 8112, endpoints: ['/health', '/v1/info', '/v1/version'] },
  'data-platform-service': { port: 8113, endpoints: ['/health', '/v1/info', '/v1/version'] },
  'compliance-ops-service': { port: 8114, endpoints: ['/health', '/v1/info', '/v1/version'] },
};

// Helper function to get random service
function getRandomService() {
  const serviceNames = Object.keys(services);
  return serviceNames[Math.floor(Math.random() * serviceNames.length)];
}

// Helper function to get random endpoint for a service
function getRandomEndpoint(service) {
  const endpoints = services[service].endpoints;
  return endpoints[Math.floor(Math.random() * endpoints.length)];
}

// Main test function
export default function () {
  const service = getRandomService();
  const endpoint = getRandomEndpoint(service);
  const port = services[service].port;
  const url = `http://localhost:${port}${endpoint}`;

  const startTime = Date.now();
  
  const response = http.get(url, {
    headers: {
      'Content-Type': 'application/json',
      'User-Agent': 'k6-performance-test',
    },
    timeout: '30s',
  });

  const endTime = Date.now();
  const responseTimeMs = endTime - startTime;
  responseTime.add(responseTimeMs);

  // Check response
  const success = check(response, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
    'response has content': (r) => r.body.length > 0,
  });

  if (!success) {
    errorRate.add(1);
    console.log(`Error: ${service}${endpoint} - Status: ${response.status}, Time: ${responseTimeMs}ms`);
  } else {
    errorRate.add(0);
  }

  // Add some think time between requests
  sleep(Math.random() * 2 + 1);
}

// Setup function (runs once before the test)
export function setup() {
  console.log('🚀 Starting Performance Test Suite');
  console.log(`📊 Testing ${Object.keys(services).length} services`);
  console.log('⏱️  Test duration: ~21 minutes');
  console.log('👥 Max concurrent users: 100');
  console.log('');
}

// Teardown function (runs once after the test)
export function teardown(data) {
  console.log('');
  console.log('✅ Performance Test Suite Complete');
  console.log('📈 Check the results above for performance metrics');
}

// Handle summary
export function handleSummary(data) {
  return {
    'performance-results.json': JSON.stringify(data, null, 2),
    'performance-results.html': generateHTMLReport(data),
    stdout: generateTextReport(data),
  };
}

// Generate HTML report
function generateHTMLReport(data) {
  return `
<!DOCTYPE html>
<html>
<head>
    <title>Performance Test Results</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f0f0f0; padding: 20px; border-radius: 5px; }
        .metric { margin: 10px 0; padding: 10px; border: 1px solid #ddd; border-radius: 3px; }
        .success { background: #d4edda; border-color: #c3e6cb; }
        .warning { background: #fff3cd; border-color: #ffeaa7; }
        .error { background: #f8d7da; border-color: #f5c6cb; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🚀 Financial Platform Performance Test Results</h1>
        <p>Generated on: ${new Date().toISOString()}</p>
    </div>
    
    <div class="metric">
        <h3>📊 Test Summary</h3>
        <p><strong>Total Requests:</strong> ${data.metrics.http_reqs.values.count}</p>
        <p><strong>Total Duration:</strong> ${(data.state.testRunDuration / 1000).toFixed(2)}s</p>
        <p><strong>Average Response Time:</strong> ${data.metrics.http_req_duration.values.avg.toFixed(2)}ms</p>
        <p><strong>95th Percentile:</strong> ${data.metrics.http_req_duration.values['p(95)'].toFixed(2)}ms</p>
        <p><strong>Error Rate:</strong> ${(data.metrics.http_req_failed.values.rate * 100).toFixed(2)}%</p>
    </div>
    
    <div class="metric">
        <h3>🎯 Threshold Results</h3>
        ${Object.entries(data.thresholds).map(([name, threshold]) => {
          const passed = threshold.ok;
          const className = passed ? 'success' : 'error';
          return `<div class="${className}">
            <strong>${name}:</strong> ${passed ? '✅ PASSED' : '❌ FAILED'}
          </div>`;
        }).join('')}
    </div>
</body>
</html>`;
}

// Generate text report
function generateTextReport(data) {
  return `
🚀 PERFORMANCE TEST RESULTS
===========================

📊 Test Summary:
- Total Requests: ${data.metrics.http_reqs.values.count}
- Total Duration: ${(data.state.testRunDuration / 1000).toFixed(2)}s
- Average Response Time: ${data.metrics.http_req_duration.values.avg.toFixed(2)}ms
- 95th Percentile: ${data.metrics.http_req_duration.values['p(95)'].toFixed(2)}ms
- Error Rate: ${(data.metrics.http_req_failed.values.rate * 100).toFixed(2)}%

🎯 Threshold Results:
${Object.entries(data.thresholds).map(([name, threshold]) => {
  const passed = threshold.ok;
  return `- ${name}: ${passed ? '✅ PASSED' : '❌ FAILED'}`;
}).join('\n')}

📈 Performance Analysis:
${data.metrics.http_req_duration.values['p(95)'] < 500 ? '✅ Performance meets requirements' : '❌ Performance below requirements'}
${data.metrics.http_req_failed.values.rate < 0.1 ? '✅ Error rate acceptable' : '❌ Error rate too high'}
`;
}
