import { check } from "k6";

// Check that the response has the expected status code.
export function checkStatus(res, expected) {
  check(res, {
    [`status is ${expected}`]: (r) => r.status === expected,
  });
}

// Check that the response is a valid JSON with data field.
export function checkJSON(res) {
  check(res, {
    "response is JSON": (r) => {
      try {
        r.json();
        return true;
      } catch (_) {
        return false;
      }
    },
  });
}

// Sleep for a random duration between min and max seconds (think time).
export function thinkTime(minSec, maxSec) {
  const duration = minSec + Math.random() * (maxSec - minSec);
  return __VU > 0 ? duration : 0;
}
