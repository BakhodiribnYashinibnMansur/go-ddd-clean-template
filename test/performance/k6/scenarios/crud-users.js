import http from "k6/http";
import { check, sleep } from "k6";
import { BASE_URL, DEFAULT_THRESHOLDS, headers } from "../lib/config.js";
import { randomPhone, randomUsername, signUp, signIn } from "../lib/auth.js";

export const options = {
  stages: [
    { duration: "30s", target: 10 },
    { duration: "1m", target: 25 },
    { duration: "2m", target: 50 },
    { duration: "2m", target: 100 },
    { duration: "1m", target: 50 },
    { duration: "30s", target: 0 },
  ],
  thresholds: DEFAULT_THRESHOLDS,
};

const API_KEY = __ENV.API_KEY || "test-api-key";
const ADMIN_PHONE = __ENV.ADMIN_PHONE || "+998901234567";
const ADMIN_PASSWORD = __ENV.ADMIN_PASSWORD || "TestPassword123!";

export function setup() {
  // Sign in as admin to get a token for CRUD operations
  const session = signIn(ADMIN_PHONE, ADMIN_PASSWORD);
  if (!session) {
    // If admin doesn't exist, create one
    signUp(ADMIN_PHONE, ADMIN_PASSWORD, "k6admin");
    return signIn(ADMIN_PHONE, ADMIN_PASSWORD);
  }
  return session;
}

export default function (session) {
  if (!session) return;

  const h = headers(session.access_token, API_KEY);

  // 1. Create a user
  const phone = randomPhone();
  const createRes = http.post(
    `${BASE_URL}/api/v1/users`,
    JSON.stringify({
      phone: phone,
      password: "NewUser123!",
      username: randomUsername(),
    }),
    { headers: h, tags: { name: "create-user" } }
  );
  check(createRes, { "create 201": (r) => r.status === 201 });

  if (createRes.status !== 201) return;
  const userId = createRes.json().data.id;
  sleep(0.3);

  // 2. List users
  const listRes = http.get(`${BASE_URL}/api/v1/users?page=1&limit=10`, {
    headers: h,
    tags: { name: "list-users" },
  });
  check(listRes, { "list 200": (r) => r.status === 200 });
  sleep(0.3);

  // 3. Get user
  const getRes = http.get(`${BASE_URL}/api/v1/users/${userId}`, {
    headers: h,
    tags: { name: "get-user" },
  });
  check(getRes, { "get 200": (r) => r.status === 200 });
  sleep(0.3);

  // 4. Update user
  const updateRes = http.patch(
    `${BASE_URL}/api/v1/users/${userId}`,
    JSON.stringify({ username: `updated_${randomUsername()}` }),
    { headers: h, tags: { name: "update-user" } }
  );
  check(updateRes, { "update 200": (r) => r.status === 200 });
  sleep(0.3);

  // 5. Delete user
  const deleteRes = http.del(`${BASE_URL}/api/v1/users/${userId}`, null, {
    headers: h,
    tags: { name: "delete-user" },
  });
  check(deleteRes, {
    "delete 200 or 204": (r) => r.status === 200 || r.status === 204,
  });
  sleep(0.5);
}
