package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

func main() {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	baseURL = strings.TrimRight(baseURL, "/")

	authToken := os.Getenv("AUTH_TOKEN")
	header := http.Header{}
	if authToken != "" {
		header.Add("Authorization", "Bearer "+authToken)
	}
	header.Add("Content-Type", "application/json")

	// dummy IDs for parameterized routes
	dummyUUID := "00000000-0000-0000-0000-000000000001"

	targets := []struct {
		Method string
		Path   string
		Body   []byte
	}{
		// Health
		{Method: "GET", Path: "/health/live"},
		{Method: "GET", Path: "/health/ready"},

		// User
		{Method: "GET", Path: "/api/v1/users/csrf-token"},
		{
			Method: "POST",
			Path:   "/api/v1/users/sign-in",
			Body:   []byte(`{"phone": "+998901234567", "password": "password123"}`),
		},
		{
			Method: "POST",
			Path:   "/api/v1/users/sign-up",
			Body:   []byte(`{"phone": "+998901234567", "password": "password123", "username": "StressUser"}`),
		},
		// {Method: "POST", Path: "/api/v1/users/refresh"}, // requires valid refresh token
		{Method: "POST", Path: "/api/v1/users/sign-out"},
		{Method: "GET", Path: "/api/v1/users/"},
		{Method: "GET", Path: "/api/v1/users/" + dummyUUID},
		{
			Method: "PATCH",
			Path:   "/api/v1/users/" + dummyUUID,
			Body:   []byte(`{"username": "StressUserUpdated"}`),
		},
		{Method: "DELETE", Path: "/api/v1/users/" + dummyUUID},

		// Session
		{Method: "GET", Path: "/api/v1/sessions/"},
		{Method: "GET", Path: "/api/v1/sessions/" + dummyUUID},
		{Method: "PATCH", Path: "/api/v1/sessions/" + dummyUUID + "/activity"},
		// {Method: "POST", Path: "/api/v1/sessions/revoke-all"}, // potentially destructive

		// Authz - Roles
		{Method: "GET", Path: "/api/v1/authz/roles"},
		{Method: "GET", Path: "/api/v1/authz/roles/" + dummyUUID},
		{
			Method: "POST",
			Path:   "/api/v1/authz/roles",
			Body:   []byte(`{"name": "stress_role"}`),
		},
		{
			Method: "PUT",
			Path:   "/api/v1/authz/roles/" + dummyUUID,
			Body:   []byte(`{"name": "stress_role_updated"}`),
		},
		{Method: "DELETE", Path: "/api/v1/authz/roles/" + dummyUUID},

		// Authz - Permissions
		{Method: "GET", Path: "/api/v1/authz/permissions"},
		{Method: "GET", Path: "/api/v1/authz/permissions/" + dummyUUID},
		{
			Method: "POST",
			Path:   "/api/v1/authz/permissions",
			Body:   []byte(`{"name": "stress_perm"}`),
		},
		{
			Method: "PUT",
			Path:   "/api/v1/authz/permissions/" + dummyUUID,
			Body:   []byte(`{"name": "stress_perm_updated"}`),
		},
		{Method: "DELETE", Path: "/api/v1/authz/permissions/" + dummyUUID},

		// Authz - Scopes
		{Method: "GET", Path: "/api/v1/authz/scopes"},
		{Method: "GET", Path: "/api/v1/authz/scopes/detail"},
		{
			Method: "POST",
			Path:   "/api/v1/authz/scopes",
			Body:   []byte(`{"path": "/stress", "method": "GET"}`),
		},
		// {Method: "DELETE", Path: "/api/v1/authz/scopes"}, // requires query params typically

		// Audit
		{Method: "GET", Path: "/api/v1/audit/logs"},
		{Method: "GET", Path: "/api/v1/audit/history"},

		// Admin
		{Method: "POST", Path: "/api/v1/admin/linter/run"},

		// Minio (Files)
		{Method: "GET", Path: "/api/v1/files/download"},
	}

	stageDuration := 10 * time.Second // Shortened for multiple endpoints
	stages := []struct {
		name string
		rate int
	}{
		{name: "Low Load", rate: 500},
		{name: "Medium Load", rate: 2000},
		{name: "High Load", rate: 5000},
	}

	attacker := vegeta.NewAttacker()

	for _, target := range targets {
		url := baseURL + target.Path
		log.Printf("\nTargeting: %s %s\n", target.Method, url)

		targeter := vegeta.NewStaticTargeter(vegeta.Target{
			Method: target.Method,
			URL:    url,
			Header: header,
			Body:   target.Body,
		})

		for _, stage := range stages {
			log.Printf("  Stage: %s (%d RPS)\n", stage.name, stage.rate)
			rate := vegeta.Rate{Freq: stage.rate, Per: time.Second}

			var metrics vegeta.Metrics
			for res := range attacker.Attack(targeter, rate, stageDuration, "Stress Test") {
				metrics.Add(res)
			}
			metrics.Close()

			log.Printf("    Mean Latency: %s, Success: %.2f%%\n", metrics.Latencies.Mean, metrics.Success*100)
		}
	}
	log.Println("\nStress test completed.")
}
