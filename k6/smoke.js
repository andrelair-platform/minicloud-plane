/**
 * k6 smoke load test for minicloud-plane.
 *
 * Usage (local):
 *   k6 run --env BASE_URL=http://localhost:8080 k6/smoke.js
 *
 * Usage (CI canary-load — covers postPromotionAnalysis window):
 *   kubectl port-forward svc/minicloud-plane 8080:8080 -n minicloud-plane-dev &
 *   k6 run -e CANARY_DURATION=5m k6/smoke.js
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

// When CANARY_DURATION is set (by the CI canary-load job), run a sustained
// constant load covering the BlueGreen postPromotionAnalysis window (4 checks
// × 30s + 60s initialDelay = ~3 min). Without sustained load, the analysis
// gates pass via `or vector(1)` with no real traffic data.
const CANARY_DURATION = __ENV.CANARY_DURATION;

export const options = CANARY_DURATION
  ? {
      stages: [
        { duration: '10s',           target: 3 },
        { duration: CANARY_DURATION, target: 3 },
        { duration: '10s',           target: 0 },
      ],
      thresholds: {
        http_req_failed:   ['rate<0.01'],
        http_req_duration: ['p(95)<500'],
        errors:            ['rate<0.01'],
      },
    }
  : {
      stages: [
        { duration: '10s', target: 3 },
        { duration: '20s', target: 3 },
        { duration: '5s',  target: 0 },
      ],
      thresholds: {
        http_req_failed:                      ['rate<0.01'],
        http_req_duration:                    ['p(95)<500'],
        'http_req_duration{endpoint:health}': ['p(95)<50'],
        errors:                               ['rate<0.01'],
      },
    };

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  // Health endpoint — primary liveness signal
  {
    const res = http.get(`${BASE_URL}/health`, { tags: { endpoint: 'health' } });
    const ok = check(res, {
      'health: status 200':  (r) => r.status === 200,
      'health: body has ok': (r) => {
        try { return JSON.parse(r.body).status === 'ok'; } catch { return false; }
      },
    });
    errorRate.add(!ok);
  }

  // Metrics endpoint — confirms instrumentation is healthy
  {
    const res = http.get(`${BASE_URL}/metrics`, { tags: { endpoint: 'metrics' } });
    check(res, { 'metrics: status 200': (r) => r.status === 200 });
    errorRate.add(res.status !== 200);
  }

  sleep(1);
}
