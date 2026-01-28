package main

import (
	"fmt"
	"os"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

func main() {
	targetURL := os.Getenv("BASE_URL")
	if targetURL == "" {
		targetURL = "http://localhost:8080/api/v1/health/ready"
	} else {
		targetURL = targetURL + "/health/ready"
	}

	fmt.Printf("Starting Spike Test against %s\n", targetURL)
	targeter := vegeta.NewStaticTargeter(vegeta.Target{
		Method: "GET",
		URL:    targetURL,
	})

	// Spike scenario: Low load -> HUGE SPIKE -> Low load recovery
	stages := []struct {
		name     string
		rate     int
		duration time.Duration
	}{
		{name: "Warmup", rate: 50, duration: 10 * time.Second},
		{name: "SPIKE", rate: 2000, duration: 20 * time.Second}, // Brief but intense spike
		{name: "Recovery", rate: 50, duration: 30 * time.Second},
	}

	for _, stage := range stages {
		fmt.Printf("Running Stage: %s (%d RPS for %s)...\n", stage.name, stage.rate, stage.duration)
		attacker := vegeta.NewAttacker()
		rate := vegeta.Rate{Freq: stage.rate, Per: time.Second}

		var metrics vegeta.Metrics
		for res := range attacker.Attack(targeter, rate, stage.duration, "Spike Test - "+stage.name) {
			metrics.Add(res)
		}
		metrics.Close()

		fmt.Printf("[%s] Finished: Mean Latency: %s, Max Latency: %s, Success: %.2f%%\n",
			stage.name, metrics.Latencies.Mean, metrics.Latencies.Max, metrics.Success*100)
	}
	fmt.Println("Spike test completed.")
}
