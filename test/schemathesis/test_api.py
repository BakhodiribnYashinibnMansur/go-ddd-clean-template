#!/usr/bin/env python3
"""
Schemathesis Test Script for Go Clean Template API

This script uses Schemathesis to automatically test the API based on OpenAPI schema.
It finds bugs like:
- 500 errors on edge cases
- Schema violations
- Validation bypasses
- Integration failures
"""

import schemathesis
from hypothesis import settings, Phase
import os

# Configuration
API_URL = os.getenv("API_URL", "http://localhost:8080/api/v1")
PROJECT_ROOT = os.getenv("PROJECT_ROOT", "/Users/mrb/Desktop/GCA")
# Prefer local file if exists, otherwise URL
SCHEMA_PATH = os.path.join(PROJECT_ROOT, "docs/swagger/swagger.yaml")
if os.path.exists(SCHEMA_PATH):
    SCHEMA_SOURCE = SCHEMA_PATH
else:
    SCHEMA_SOURCE = f"{API_URL}/swagger/doc.json"

print(f"Using Schema: {SCHEMA_SOURCE}")
print(f"Using API: {API_URL}")

# Create schema object
if os.path.exists(SCHEMA_SOURCE):
    schema = schemathesis.openapi.from_path(SCHEMA_SOURCE).configure(base_url=API_URL)
else:
    schema = schemathesis.openapi.from_url(SCHEMA_SOURCE).configure(base_url=API_URL)

# Configure Hypothesis settings for more thorough testing
settings.register_profile(
    "thorough",
    max_examples=100,  # Number of test cases per endpoint
    deadline=5000,     # 5 seconds timeout per test
    phases=[Phase.explicit, Phase.reuse, Phase.generate, Phase.target],
)
settings.load_profile("thorough")


# Test all endpoints with automatic validation
@schema.parametrize()
@settings(max_examples=50)
def test_api(case):
    """
    Test each API endpoint with automatically generated test data.
    
    This will:
    1. Generate random valid and invalid inputs
    2. Send requests to the API
    3. Validate responses against the schema
    4. Check for common issues (500 errors, schema violations, etc.)
    """
    response = case.call()
    
    # Validate response against schema
    case.validate_response(response)
    
    # Additional custom checks
    if response.status_code >= 500:
        print(f"Server error: {response.status_code} for {case.method} {case.path}")


# Stateful testing - tests realistic workflows
APIWorkflow = schema.as_state_machine()


class TestAPIStateful(APIWorkflow.TestCase):
    """
    Stateful testing class that tests realistic API workflows.
    
    Example workflows:
    - Create user -> Get user -> Update user -> Delete user
    - Sign in -> Create session -> Get session -> Revoke session
    - Create policy -> Assign to role -> Test authorization
    """
    
    def setup_method(self):
        """Setup before each test"""
        # You can add authentication tokens here
        self.auth_token = None
    
    def teardown_method(self):
        """Cleanup after each test"""
        pass


import requests

# Authentication Configuration
LOGIN = os.getenv("TEST_LOGIN", "bakhodiryashinmansur@gmail.com")
PASSWORD = os.getenv("TEST_PASSWORD", "0224")
TOKEN = None


def get_token():
    """Retrieve authentication token by signing in"""
    global TOKEN
    if TOKEN:
        return TOKEN
        
    try:
        url = f"{API_URL}/auth/sign-in"
        payload = {
            "login": LOGIN,
            "password": PASSWORD,
            "session": {
                "device_name": "SchemathesisRunner-Py",
                "device_type": "BOT",
                "ip_address": "127.0.0.1",
                "user_agent": "Schemathesis/v2-py",
                "os": "MacOS",
                "os_version": "14.0",
                "browser": "Python",
                "browser_version": "3.11"
            }
        }
        
        response = requests.post(url, json=payload, timeout=5)
        if response.status_code == 200:
            data = response.json().get("data", {})
            # Adjust based on actual response structure
            TOKEN = data.get("access_token") or data.get("token")
            if TOKEN:
                print(f"✅ Authenticated as {LOGIN}")
            else:
                 print(f"⚠️  Authentication successful but no token found in response: {data.keys()}")
        else:
            print(f"⚠️  Authentication failed: {response.status_code} - {response.text}")
            
    except Exception as e:
        print(f"⚠️  Authentication error: {e}")
        
    return TOKEN


# Custom hooks for authentication
@schema.hooks.before_call
def add_auth_header(context, case):
    """Add authentication header to requests that need it"""
    # Skip auth endpoints to avoid infinite loops or unnecessary auth
    if "/auth/" in case.path and "sign-out" not in case.path:
        return
    
    token = get_token()
    if token:
        case.headers = case.headers or {}
        case.headers["Authorization"] = f"Bearer {token}"


@schema.hooks.after_call
def log_failures(context, case, response):
    """Log failed requests for debugging"""
    if response.status_code >= 400:
        # Check if 400 is expected (Schema compliant but logic error)
        # or if 500 (Server error)
        pass 


if __name__ == "__main__":
    # Run tests using pytest
    import pytest
    import sys
    
    # Run with verbose output
    sys.exit(pytest.main([__file__, "-v", "--tb=short"]))
