#!/bin/bash
set -e

# Configuration
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
REPORT_DIR="$PROJECT_ROOT/docs/report/schemathesis"
REPORT_FILE="$REPORT_DIR/project_report.txt"
VENV_DIR="$PROJECT_ROOT/.venv"
SCHEMATHESIS_BIN="$VENV_DIR/bin/schemathesis"
API_URL="${API_URL:-http://localhost:8080/api/v1}"
SCHEMA_PATH="${SCHEMA_PATH:-$PROJECT_ROOT/docs/swagger/swagger.yaml}"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

mkdir -p "$REPORT_DIR"

echo -e "${GREEN}🧪 Starting Schemathesis Tests${NC}"
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
USERNAME="${USERNAME:-admin}"
PASSWORD="${PASSWORD:-admin}"

TOKEN_RESP=$(curl -s -X POST "$API_URL/auth/sign-in" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")

# Extract token (trying jq first, then python fallback)
TOKEN=""
if command -v jq &> /dev/null; then
    TOKEN=$(echo "$TOKEN_RESP" | jq -r '.data.access_token')
else
    # Python fallback for extracting token
    TOKEN=$(echo "$TOKEN_RESP" | python3 -c "import sys, json; print(json.load(sys.stdin).get('data', {}).get('access_token', ''))")
fi

ARGS=()
ARGS+=(run "$SCHEMA_PATH")
ARGS+=(--url="$API_URL")
ARGS+=(--checks=all)
ARGS+=(--max-examples=100)
ARGS+=(--show-errors-tracebacks)
ARGS+=(--workers=4)
# Exclude problematic endpoint
ARGS+=(--exclude-path "/authz/permissions/{perm_id}/scopes") 

if [ -n "$TOKEN" ] && [ "$TOKEN" != "null" ]; then
    echo -e "${GREEN}✅ Authenticated successfully!${NC}"
    ARGS+=(--header "Authorization: Bearer $TOKEN")
else
    echo -e "${RED}❌ Authentication failed. Response: $TOKEN_RESP${NC}"
    echo "Continuing without authentication..."
fi

# 3. Run Tests
echo -e "${YELLOW}🚀 Running tests...${NC}"
"$SCHEMATHESIS_BIN" "${ARGS[@]}" 2>&1 | tee "$REPORT_FILE"
