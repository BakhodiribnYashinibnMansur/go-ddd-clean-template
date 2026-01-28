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

	fmt.Printf("Starting Breakpoint (Capacity) Test against %s\n", targetURL)
	fmt.Println("Goal: Find the maximum RPS where success rate stays > 99% and Latency P99 < 1s")

	targeter := vegeta.NewStaticTargeter(vegeta.Target{
		Method: "GET",
		URL:    targetURL,
	})

	startRate := 100
	maxRate := 5000
	step := 100
	durationPerStep := 10 * time.Second

	attacker := vegeta.NewAttacker()

	for currentRate := startRate; currentRate <= maxRate; currentRate += step {
		fmt.Printf("Testing at %d RPS...\n", currentRate)

		rate := vegeta.Rate{Freq: currentRate, Per: time.Second}
		var metrics vegeta.Metrics

		for res := range attacker.Attack(targeter, rate, durationPerStep, "Breakpoint Test") {
			metrics.Add(res)
		}
		metrics.Close()

		fmt.Printf("-> %d RPS: Success: %.2f%%, P99: %s\n", currentRate, metrics.Success*100, metrics.Latencies.P99)

		if metrics.Success < 0.99 || metrics.Latencies.P99 > 1*time.Second {
			fmt.Printf("\nFAILED criteria at %d RPS.\n", currentRate)
			fmt.Printf("Max Sustainable RPS: ~%d\n", currentRate-step)
			return
		}
	}

	fmt.Printf("Reached max tested rate of %d RPS without breaking criteria.\n", maxRate)
}
