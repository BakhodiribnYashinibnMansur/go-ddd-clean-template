import http from "k6/http";
import { check, sleep } from "k6";
import { BASE_URL, DEFAULT_THRESHOLDS, headers } from "../lib/config.js";
import { randomPhone, signUp, signIn, signOut } from "../lib/auth.js";

export const options = {
  scenarios: {
    // 70% reads — list/get operations
    reads: {
      executor: "ramping-vus",
      startVUs: 0,
      stages: [
        { duration: "1m", target: 20 },
        { duration: "5m", target: 50 },
        { duration: "2m", target: 70 },
        { duration: "2m", target: 0 },
      ],
      exec: "readScenario",
    },
    // 20% writes — create/update/delete
    writes: {
      executor: "constant-vus",
      vus: 15,
      duration: "10m",
      exec: "writeScenario",
    },
    // 10% auth — sign-up/sign-in/sign-out cycles
    auth: {
      executor: "constant-arrival-rate",
      rate: 5,
      timeUnit: "1s",
      duration: "10m",
      preAllocatedVUs: 10,
      maxVUs: 20,
      exec: "authScenario",
    },
  },
  thresholds: DEFAULT_THRESHOLDS,
};

const API_KEY = __ENV.API_KEY || "test-api-key";
const PASSWORD = "TestPassword123!";

// Shared session created in setup.
export function setup() {
  const phone = randomPhone();
  signUp(phone, PASSWORD);
  return signIn(phone, PASSWORD);
}

// Read scenario — list endpoints.
export function readScenario(session) {
  if (!session) return;
  const h = headers(session.access_token, API_KEY);

  const endpoints = [
    "/api/v1/users?page=1&limit=10",
    "/api/v1/sessions?page=1&limit=10",
    "/api/v1/roles?page=1&limit=10",
    "/api/v1/notifications?page=1&limit=10",
    "/api/v1/translations?page=1&limit=10",
    "/api/v1/announcements?page=1&limit=10",
    "/api/v1/site-settings?page=1&limit=10",
  ];

  const endpoint = endpoints[Math.floor(Math.random() * endpoints.length)];
  const res = http.get(`${BASE_URL}${endpoint}`, {
    headers: h,
    tags: { name: "read" },
  });
  check(res, { "read 200": (r) => r.status === 200 });
  sleep(0.5 + Math.random());
}

// Write scenario — create and delete notifications.
export function writeScenario(session) {
  if (!session) return;
  const h = headers(session.access_token, API_KEY);

  const createRes = http.post(
    `${BASE_URL}/api/v1/notifications`,
    JSON.stringify({
      title: `k6 test ${Date.now()}`,
      body: "Load test notification",
      type: "INFO",
    }),
    { headers: h, tags: { name: "write-create" } }
  );
  check(createRes, { "write create": (r) => r.status === 201 || r.status === 200 });

  if (createRes.status === 201 || createRes.status === 200) {
    const body = createRes.json();
    if (body && body.data && body.data.id) {
      sleep(0.3);
      const delRes = http.del(
        `${BASE_URL}/api/v1/notifications/${body.data.id}`,
        null,
        { headers: h, tags: { name: "write-delete" } }
      );
      check(delRes, { "write delete": (r) => r.status === 200 || r.status === 204 });
    }
  }
  sleep(1 + Math.random() * 2);
}

// Auth scenario — full auth cycle.
export function authScenario() {
  const phone = randomPhone();
  signUp(phone, PASSWORD);
  sleep(0.3);

  const session = signIn(phone, PASSWORD);
  if (session) {
    sleep(0.5);
    signOut(session.access_token);
  }
  sleep(0.5);
}
