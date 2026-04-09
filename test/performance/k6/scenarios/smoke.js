import http from "k6/http";
import { check, sleep } from "k6";
import { BASE_URL, SMOKE_THRESHOLDS, headers } from "../lib/config.js";
import { randomPhone, signUp, signIn, signOut } from "../lib/auth.js";

export const options = {
  vus: 1,
  duration: "30s",
  thresholds: SMOKE_THRESHOLDS,
};

const API_KEY = __ENV.API_KEY || "test-api-key";
const PASSWORD = "TestPassword123!";

export default function () {
  // Health check
  const health = http.get(`${BASE_URL}/health`);
  check(health, { "health 200": (r) => r.status === 200 });

  // Quick auth flow
  const phone = randomPhone();
  signUp(phone, PASSWORD);

  const session = signIn(phone, PASSWORD);
  if (session) {
    // Authenticated request
    const users = http.get(`${BASE_URL}/api/v1/users`, {
      headers: headers(session.access_token, API_KEY),
    });
    check(users, { "users list 200": (r) => r.status === 200 });

    signOut(session.access_token);
  }

  sleep(1);
}
