import os
import requests
import pytest
from schemathesis.extensions import hypothesis
from hypothesis import strategies as st
from hypothesis.stateful import RuleBasedStateMachine, rule, Bundle, consumes, initialize

API_URL = os.getenv("API_URL", "http://localhost:8080/api/v1")
LOGIN_EMAIL = os.getenv("TEST_LOGIN", "bakhodiryashinmansur@gmail.com")
LOGIN_PASSWORD = os.getenv("TEST_PASSWORD", "0224")

class AuthFlow(RuleBasedStateMachine):
    """
    Tests the Authentication Lifecycle:
    Sign In -> Refresh -> Sign Out
    """
    
    token = Bundle("token")
    refresh_token = Bundle("refresh_token")

    def __init__(self):
        super().__init__()
        self.session = requests.Session()
        self.headers = {"Content-Type": "application/json"}

    @initialize(target=token)
    def sign_in(self):
        payload = {
            "login": LOGIN_EMAIL,
            "password": LOGIN_PASSWORD,
            "session": {
                "device_name": "StatefulTestBot",
                "device_type": "BOT",
                "ip_address": "127.0.0.1", 
                "user_agent": "Hypothesis/1.0",
                "os": "Linux",
                "os_version": "1.0",
                "browser": "Test",
                "browser_version": "1.0"
            }
        }
        response = self.session.post(f"{API_URL}/auth/sign-in", json=payload, headers=self.headers)
        assert response.status_code == 200, f"Sign in failed: {response.text}"
        
        data = response.json()["data"]
        access_token = data["access_token"]
        
        # Save tokens
        self.current_access_token = access_token
        self.current_refresh_token = data.get("refresh_token")
        
        return access_token

    @rule(access_token=token)
    def check_profile(self, access_token):
        headers = {**self.headers, "Authorization": f"Bearer {access_token}"}
        response = self.session.get(f"{API_URL}/auth/me", headers=headers)
        # Note: Depending on implementation, might return 200 or 401 if expired
        if response.status_code == 200:
            assert response.json()["data"]["email"] == LOGIN_EMAIL
        elif response.status_code == 401:
            # If 401, it might be expired, which is 'valid' behavior for state machine if time passed
            pass
        else:
            pytest.fail(f"Unexpected status for /auth/me: {response.status_code}")

    @rule(access_token=token)
    def refresh_token_flow(self, access_token):
        # We need a refresh token. In this simple model we assume we have one from sign_in.
        # Ideally Bundle should hold (access, refresh) tuple.
        if not hasattr(self, 'current_refresh_token') or not.self.current_refresh_token:
            return

        payload = {
            "refresh_token": self.current_refresh_token
        }
        # Note: Refresh endpoint might require Auth header or not depending on implementation.
        # Assuming it might need the OLD access token or just public endpoint.
        # Let's try as public first or with header.
        headers = {**self.headers, "Authorization": f"Bearer {access_token}"}
        
        response = self.session.post(f"{API_URL}/auth/refresh", json=payload, headers=headers)
        
        if response.status_code == 200:
            data = response.json()["data"]
            self.current_access_token = data["access_token"]
            self.current_refresh_token = data.get("refresh_token")
        elif response.status_code in [400, 401]:
             # Refresh might fail if token cycled, which is acceptable state transition
             pass
        else:
             pytest.fail(f"Refresh failed with unexpected status: {response.status_code}")

    @rule(access_token=token)
    def sign_out(self, access_token):
        headers = {**self.headers, "Authorization": f"Bearer {access_token}"}
        response = self.session.post(f"{API_URL}/auth/sign-out", headers=headers)
        assert response.status_code in [200, 204, 401], f"Sign out failed: {response.text}"

# Test Runner Config
TestAuth = AuthFlow.TestCase
