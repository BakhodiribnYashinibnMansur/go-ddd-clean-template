import http from "k6/http";
import { check, sleep } from "k6";
import { BASE_URL, DEFAULT_THRESHOLDS, headers } from "../lib/config.js";
import { signUp, signIn, randomPhone } from "../lib/auth.js";

export const options = {
  stages: [
    { duration: "30s", target: 5 },
    { duration: "2m", target: 10 },
    { duration: "2m", target: 20 },
    { duration: "1m", target: 10 },
    { duration: "30s", target: 0 },
  ],
  thresholds: {
    ...DEFAULT_THRESHOLDS,
    "http_req_duration{name:upload}": ["p(95)<2000"],
    "http_req_duration{name:download}": ["p(95)<1000"],
  },
};

const API_KEY = __ENV.API_KEY || "test-api-key";
const PASSWORD = "TestPassword123!";

// Minimal 1x1 pixel GIF (43 bytes).
const GIF_BYTES = new Uint8Array([
  0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00, 0x80, 0x00,
  0x00, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x21, 0xf9, 0x04, 0x00, 0x00,
  0x00, 0x00, 0x00, 0x2c, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00,
  0x00, 0x02, 0x02, 0x44, 0x01, 0x00, 0x3b,
]);

export function setup() {
  const phone = randomPhone();
  signUp(phone, PASSWORD);
  return signIn(phone, PASSWORD);
}

export default function (session) {
  if (!session) return;

  const h = { Authorization: `Bearer ${session.access_token}`, "X-API-Key": API_KEY };

  // 1. Upload image
  const file = http.file(GIF_BYTES.buffer, "test.gif", "image/gif");
  const uploadRes = http.post(`${BASE_URL}/api/v1/files/upload/image`, { file: file }, {
    headers: h,
    tags: { name: "upload" },
  });
  check(uploadRes, { "upload success": (r) => r.status === 200 || r.status === 201 });

  if (uploadRes.status !== 200 && uploadRes.status !== 201) {
    sleep(1);
    return;
  }

  sleep(0.5);

  // 2. List files
  const listRes = http.get(`${BASE_URL}/api/v1/files?page=1&limit=10`, {
    headers: { ...h, "Content-Type": "application/json" },
    tags: { name: "list-files" },
  });
  check(listRes, { "list files 200": (r) => r.status === 200 });

  sleep(1);
}
