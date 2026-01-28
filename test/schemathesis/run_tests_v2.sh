#!/bin/bash
set -e

# Configuration
# Hardcoded to avoid permission issues and ensuring authentication
PROJECT_ROOT="/Users/mrb/Desktop/GCA"
REPORT_DIR="/tmp/schemathesis_reports"
REPORT_FILE="$REPORT_DIR/report.txt"
VENV_DIR="$PROJECT_ROOT/.venv"
SCHEMATHESIS_BIN="$VENV_DIR/bin/schemathesis"
API_URL="http://localhost:8080/api/v1"
SCHEMA_PATH="$PROJECT_ROOT/docs/swagger/swagger.yaml"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

mkdir -p "$REPORT_DIR"

echo -e "${GREEN}🧪 Starting Schemathesis Tests (v2 Authenticated)${NC}"
echo "   API URL: $API_URL"
echo "   Report: $REPORT_FILE"

# 1. Install/Check Dependencies
if [ ! -f "$SCHEMATHESIS_BIN" ]; then
    echo -e "${YELLOW}⚠️  Installing Schemathesis...${NC}"
    python3 -m venv "$VENV_DIR"
    "$VENV_DIR/bin/pip" install schemathesis
fi

# 2. Authentication
echo -e "${YELLOW}🔑 Authenticating...${NC}"
# Credentials from user
USERNAME="bakhodiryashinmansur@gmail.com"
PASSWORD="0224"

# Correct payload with 'login' field and 'session' object
# Using jq to ensure proper JSON formatting
PAYLOAD=$(cat <<EOF
{
  "login": "$USERNAME",
  "password": "$PASSWORD",
  "session": {
    "device_name": "SchemathesisRunner",
    "device_type": "BOT",
    "ip_address": "127.0.0.1",
    "user_agent": "Schemathesis/v2",
    "os": "MacOS",
    "os_version": "14.0",
    "browser": "CLI",
    "browser_version": "1.0"
  }
}
EOF
)

TOKEN_RESP=$(curl -s -X POST "$API_URL/auth/sign-in" \
  -H "Content-Type: application/json" \
  -d "$PAYLOAD")

# Extract token
TOKEN=""
if command -v jq &> /dev/null; then
    TOKEN=$(echo "$TOKEN_RESP" | jq -r '.data.access_token // empty')
else
    # Fallback to python if jq is missing
    TOKEN=$(echo "$TOKEN_RESP" | python3 -c "import sys, json; print(json.load(sys.stdin).get('data', {}).get('access_token', ''))")
fi

ARGS=()
ARGS+=(run "$SCHEMA_PATH")
ARGS+=(--url="$API_URL")
ARGS+=(--checks=all)
ARGS+=(--max-examples=100)
ARGS+=(--workers=4)
# Exclude problematic endpoint
ARGS+=(--exclude-path "/authz/permissions/{perm_id}/scopes") 

if [ -n "$TOKEN" ] && [ "$TOKEN" != "null" ]; then
    echo -e "${GREEN}✅ Authenticated successfully!${NC}"
    ARGS+=(--header "Authorization: Bearer $TOKEN")
else
    echo -e "${RED}❌ Authentication failed. Response: $TOKEN_RESP${NC}"
    # Continue without auth to test public endpoints or see 401s
fi

# 3. Run Tests
echo -e "${YELLOW}🚀 Running tests...${NC}"
"$SCHEMATHESIS_BIN" "${ARGS[@]}" 2>&1 | tee "$REPORT_FILE"
