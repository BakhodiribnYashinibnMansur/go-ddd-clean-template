import http from "k6/http";
import { check, sleep } from "k6";
import { BASE_URL, DEFAULT_THRESHOLDS, headers } from "../lib/config.js";
import { randomPhone, signUp, signIn, signOut } from "../lib/auth.js";

export const options = {
  stages: [
    { duration: "30s", target: 10 },
    { duration: "1m", target: 25 },
    { duration: "2m", target: 50 },
    { duration: "1m", target: 25 },
    { duration: "30s", target: 0 },
  ],
  thresholds: {
    ...DEFAULT_THRESHOLDS,
    "http_req_duration{name:sign-in}": ["p(95)<800"],
    "http_req_duration{name:profile}": ["p(95)<500"],
  },
};

const API_KEY = __ENV.API_KEY || "test-api-key";
const PASSWORD = "TestPassword123!";

export default function () {
  // 1. Sign up with a unique phone
  const phone = randomPhone();
  signUp(phone, PASSWORD);
  sleep(0.5);

  // 2. Sign in
  const session = signIn(phone, PASSWORD);
  if (!session) return;
  sleep(0.5);

  // 3. Get own profile
  const profile = http.get(`${BASE_URL}/api/v1/users/${session.user_id}`, {
    headers: headers(session.access_token, API_KEY),
    tags: { name: "profile" },
  });
  check(profile, { "profile 200": (r) => r.status === 200 });
  sleep(0.5);

  // 4. Sign out
  signOut(session.access_token);
  sleep(1);
}
