// Shared configuration for k6 load tests.
export const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";

// Default thresholds applied to all scenarios unless overridden.
export const DEFAULT_THRESHOLDS = {
  http_req_duration: ["p(95)<500", "p(99)<1000"],
  http_req_failed: ["rate<0.01"],
};

// Strict thresholds for smoke tests.
export const SMOKE_THRESHOLDS = {
  http_req_duration: ["p(95)<300", "p(99)<500"],
  http_req_failed: ["rate<0.001"],
};

// Common HTTP headers.
export function headers(token, apiKey) {
  const h = { "Content-Type": "application/json" };
  if (token) h["Authorization"] = `Bearer ${token}`;
  if (apiKey) h["X-API-Key"] = apiKey;
  return h;
}
