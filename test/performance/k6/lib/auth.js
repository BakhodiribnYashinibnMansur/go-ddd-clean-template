import http from "k6/http";
import { check } from "k6";
import { BASE_URL, headers } from "./config.js";

const API_KEY = __ENV.API_KEY || "test-api-key";

// Generate a random Uzbek phone number.
export function randomPhone() {
  const prefix = "99890";
  const suffix = Math.floor(1000000 + Math.random() * 9000000).toString();
  return `+${prefix}${suffix}`;
}

// Generate a random username.
export function randomUsername() {
  return `k6user_${Date.now()}_${Math.floor(Math.random() * 100000)}`;
}

// Sign up a new user. Returns the parsed response body.
export function signUp(phone, password, username) {
  const payload = JSON.stringify({
    phone: phone,
    password: password,
    username: username || randomUsername(),
  });

  const res = http.post(`${BASE_URL}/api/v1/auth/sign-up`, payload, {
    headers: headers(null, API_KEY),
  });

  check(res, { "sign-up status 201": (r) => r.status === 201 });
  return res;
}

// Sign in with credentials. Returns { access_token, user_id, session_id } or null.
export function signIn(login, password) {
  const payload = JSON.stringify({
    login: login,
    password: password,
    device_type: "WEB",
  });

  const res = http.post(`${BASE_URL}/api/v1/auth/sign-in`, payload, {
    headers: headers(null, API_KEY),
  });

  check(res, { "sign-in status 200": (r) => r.status === 200 });

  if (res.status === 200) {
    const body = res.json();
    return {
      access_token: body.data.access_token,
      user_id: body.data.user_id,
      session_id: body.data.session_id,
      refresh_token: body.data.refresh_token,
    };
  }
  return null;
}

// Sign out the current session.
export function signOut(token) {
  const res = http.post(`${BASE_URL}/api/v1/auth/sign-out`, null, {
    headers: headers(token, API_KEY),
  });

  check(res, { "sign-out status 200": (r) => r.status === 200 });
  return res;
}
